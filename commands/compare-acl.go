package commands

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-app-db/db"
	lib "github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/config"
)

var CompareACLCmd = CompareACL{
	command: command{
		name:        "compare-acl",
		description: "Compares the access permissions in the configurated set of access controllers to an access control list in a database",
		usage:       "[--with-pin] --dsn <DSN> [--table:ACL <table>] [-table:audit <table>] [--file <file>]",

		dsn: "",
		tables: tables{
			ACL:   "ACL",
			Audit: "",
		},
		withPIN:  false,
		lockfile: "",
		config:   config.DefaultConfig,
	},

	file:  "",
	debug: false,

	template: `ACL DIFF REPORT {{ .DateTime }}
{{range $id,$value := .Diffs}}
  DEVICE {{ $id }}{{if or $value.Updated $value.Added $value.Deleted}}{{else}} OK{{end}}{{if $value.Updated}}
    Incorrect:  {{range $value.Updated}}{{.}}
                {{end}}{{end}}{{if $value.Added}}
    Missing:    {{range $value.Added}}{{.}}
                {{end}}{{end}}{{if $value.Deleted}}
    Unexpected: {{range $value.Deleted}}{{.}}
                {{end}}{{end}}{{end}}
`,
}

type CompareACL struct {
	command
	file     string
	template string
	debug    bool
}

func (cmd *CompareACL) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s [--debug] [--config <file>] compare-acl [--with-pin] [--file <file>] --dsn <DSN> [--table:ACL <table>] [-table:audit <table>]\n", APP)
	fmt.Println()
	fmt.Println("  Compares the access permissions in the configurated set of access controllers to an access control list in a database")
	fmt.Println()

	helpOptions(cmd.FlagSet())

	fmt.Println()
	fmt.Println("  Examples:")
	fmt.Println(`    uhppote-app-db --debug compare-acl --with-pin --dsn "sqlite3://./db/ACL.db"`)
	fmt.Println(`    uhppote-app-db --debug compare-acl --with-pin --dsn "sqlite3://./db/ACL.db" --table:ACL ACL2 --table:audit AuditTrail`)
	fmt.Println()
}

func (cmd *CompareACL) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("compare-acl", flag.ExitOnError)

	flagset.StringVar(&cmd.dsn, "dsn", cmd.dsn, "DSN for database")
	flagset.StringVar(&cmd.tables.ACL, "table:ACL", cmd.tables.ACL, "ACL table name. Defaults to ACL")
	flagset.StringVar(&cmd.tables.Audit, "table:audit", cmd.tables.Audit, "Audit trail table name. Defaults to ''")
	flagset.BoolVar(&cmd.withPIN, "with-pin", cmd.withPIN, "Include card keypad PIN code when comparing access controllers")
	flagset.StringVar(&cmd.file, "file", cmd.file, "Optional filepath for compare report. Defaults to stdout")
	flagset.StringVar(&cmd.lockfile, "lockfile", cmd.lockfile, "Filepath for lock file. Defaults to <tmp>/uhppoted-app-db.lock")

	return flagset
}

func (cmd *CompareACL) Execute(args ...any) error {
	options := args[0].(*Options)

	cmd.config = options.Config
	cmd.debug = options.Debug

	// ... check parameters
	if strings.TrimSpace(cmd.dsn) == "" {
		return fmt.Errorf("invalid database DSN")
	}

	if strings.TrimSpace(cmd.tables.ACL) == "" {
		return fmt.Errorf("invalid ACL table")
	}

	// ... locked?
	if kraken, err := lock(cmd.lockfile); err != nil {
		return err
	} else {
		defer func() {
			infof("compare-acl", "Removing lockfile")
			kraken.Release()
		}()
	}

	// ... get config
	conf := config.NewConfig()
	if err := conf.Load(cmd.config); err != nil {
		return fmt.Errorf("could not load configuration (%v)", err)
	}

	u, devices := getDevices(conf, cmd.debug)

	// ... retrieve ACL from DB
	f := func(table lib.Table, devices []uhppote.Device) (*lib.ACL, []error, error) {
		if cmd.withPIN {
			return lib.ParseTable(&table, devices, false)
		} else {
			return lib.ParseTable(&table, devices, false)
		}
	}

	if table, err := getACL(cmd.dsn, cmd.tables.ACL, cmd.withPIN); err != nil {
		return err
	} else if acl, warnings, err := f(table, devices); err != nil {
		return err
	} else if acl == nil {
		return fmt.Errorf("error creating ACL from DB table (%v)", acl)
	} else {
		if cmd.debug {
			acl.Print(os.Stdout)
		}

		for _, w := range warnings {
			warnf("compare-acl", "%v", w)
		}

		diff, err := cmd.compare(u, devices, *acl)
		if err != nil {
			return err
		}

		bytes, err := cmd.format(diff)
		if err != nil {
			return err
		}

		if cmd.tables.Audit != "" {
			recordset := diff2recordset(diff, cmd.withPIN)
			if err := stash("compare", cmd.dsn, cmd.tables.Audit, recordset); err != nil {
				return err
			}
		}

		if cmd.file != "" {
			if err := os.MkdirAll(filepath.Dir(cmd.file), 0750); err != nil {
				return err
			} else if err := os.WriteFile(cmd.file, bytes, 0660); err != nil {
				return err
			}
		} else if _, err := fmt.Printf("%v", string(bytes)); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *CompareACL) compare(u uhppote.IUHPPOTE, devices []uhppote.Device, acl lib.ACL) (lib.SystemDiff, error) {
	current, errors := lib.GetACL(u, devices)
	if len(errors) > 0 {
		return lib.SystemDiff{}, fmt.Errorf("%v", errors)
	}

	f := func() (map[uint32]lib.Diff, error) {
		if cmd.withPIN {
			return lib.CompareWithPIN(current, acl)
		} else {
			return lib.Compare(current, acl)
		}
	}

	if d, err := f(); err != nil {
		return lib.SystemDiff{}, err
	} else {
		return lib.SystemDiff(d), nil
	}
}

func (cmd *CompareACL) format(diff map[uint32]lib.Diff) ([]byte, error) {
	var b bytes.Buffer

	t, err := template.New("report").Parse(cmd.template)
	if err != nil {
		return nil, err
	}

	rpt := struct {
		DateTime string
		Diffs    map[uint32]lib.Diff
	}{
		DateTime: time.Now().Format("2006-01-02 15:04:05"),
		Diffs:    diff,
	}

	if err := t.Execute(&b, rpt); err != nil {
		return nil, err
	} else {
		return b.Bytes(), nil
	}
}

func diff2recordset(diff lib.SystemDiff, withPIN bool) []db.AuditRecord {
	now := time.Now()
	recordset := []db.AuditRecord{}

	auditRecord := func(controller uint32, card core.Card, status string) db.AuditRecord {
		return db.AuditRecord{
			Timestamp:  now,
			Operation:  "compare",
			Controller: controller,
			CardNumber: card.CardNumber,
			Status:     status,
			Card:       format(card, withPIN),
		}
	}

	for controller, v := range diff {
		for _, card := range v.Updated {
			recordset = append(recordset, auditRecord(controller, card, "incorrect"))
		}

		for _, card := range v.Added {
			recordset = append(recordset, auditRecord(controller, card, "missing"))
		}

		for _, card := range v.Deleted {
			recordset = append(recordset, auditRecord(controller, card, "extra"))
		}
	}

	return recordset
}

func format(c core.Card, withPIN bool) string {
	door := func(door uint8, access uint8) string {
		switch access {
		case 0:
			return "N"
		case 1:
			return "Y"
		default:
			return fmt.Sprintf("%v", access)
		}
	}

	if withPIN {
		return fmt.Sprintf("%-10v %-10s %-10s %-3v %-3v %-3v %-3v %-5v",
			c.CardNumber,
			c.From,
			c.To,
			door(1, c.Doors[1]),
			door(2, c.Doors[2]),
			door(3, c.Doors[3]),
			door(4, c.Doors[4]),
			c.PIN)
	} else {
		return fmt.Sprintf("%-10v %-10s %-10s %-3v %-3v %-3v %-3v",
			c.CardNumber,
			c.From,
			c.To,
			door(1, c.Doors[1]),
			door(2, c.Doors[2]),
			door(3, c.Doors[3]),
			door(4, c.Doors[4]))
	}
}
