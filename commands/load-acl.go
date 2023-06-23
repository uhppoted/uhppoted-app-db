package commands

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-app-db/db"
	lib "github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/config"
)

var LoadACLCmd = LoadACL{
	command: command{
		name:        "load-acl",
		description: "Retrieves an access control list from a database and updates the configured set of access controllers",
		usage:       "[--with-pin] --dsn <DSN> [--table:ACL <table>] [-table:audit <table>] [-table:log <table>]",

		dsn: "",
		tables: tables{
			ACL:   "ACL",
			Audit: "",
		},
		withPIN:  false,
		lockfile: "",
		config:   config.DefaultConfig,
		debug:    false,
	},
}

type LoadACL struct {
	command
}

func (cmd *LoadACL) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s [--debug] [--config <file>] load-acl [--with-pin] --dsn <DSN> [--table:ACL <table>] [--table:ACL <table>] [-table:log <table>]\n", APP)
	fmt.Println()
	fmt.Println("  Retrieves an access control list from a database and updates the configured set of access controllers")
	fmt.Println()

	helpOptions(cmd.FlagSet())

	fmt.Println()
	fmt.Println("  Examples:")
	fmt.Println(`    uhppote-app-db --debug load-acl --with-pin --dsn "sqlite3://./db/ACL.db"`)
	fmt.Println(`    uhppote-app-db --debug load-acl --with-pin --dsn "sqlite3://./db/ACL.db" --table:ACL ACL --table:audit AuditTrail --table:log OpsLog`)
	fmt.Println()
}

func (cmd *LoadACL) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("load-acl", flag.ExitOnError)

	flagset.StringVar(&cmd.dsn, "dsn", cmd.dsn, "DSN for database")
	flagset.StringVar(&cmd.tables.ACL, "table:ACL", cmd.tables.ACL, "ACL table name. Defaults to ACL")
	flagset.StringVar(&cmd.tables.Audit, "table:audit", cmd.tables.Audit, "Audit trail table name. Defaults to ''")
	flagset.StringVar(&cmd.tables.Log, "table:log", cmd.tables.Log, "Operations log table name. Defaults to ''")
	flagset.BoolVar(&cmd.withPIN, "with-pin", cmd.withPIN, "Include card keypad PIN code when updating access controllers")
	flagset.StringVar(&cmd.lockfile, "lockfile", cmd.lockfile, "Filepath for lock file. Defaults to <tmp>/uhppoted-app-db.lock")

	return flagset
}

func (cmd *LoadACL) Execute(args ...any) error {
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
			infof("load-acl", "Removing lockfile")
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

		report, errors := cmd.load(u, *acl)
		if len(errors) > 0 {
			return fmt.Errorf("%v", errors)
		}

		for _, w := range warnings {
			if duplicate, ok := w.(*lib.DuplicateCardError); ok {
				for k, v := range report {
					v.Errored = append(v.Errored, duplicate.CardNumber)
					report[k] = v
				}
			}
		}

		if cmd.tables.Audit != "" {
			recordset := report2audit(report)
			if err := stashToAudit(cmd.dsn, cmd.tables.Audit, recordset); err != nil {
				return err
			}
		}

		if cmd.tables.Log != "" {
			recordset := report2log(report)
			if err := stashToLog(cmd.dsn, cmd.tables.Log, recordset); err != nil {
				return err
			}
		}

		summary := lib.Summarize(report)
		format := "%v  unchanged:%v  updated:%v  added:%v  deleted:%v  failed:%v  errors:%v"
		for _, v := range summary {
			infof("load-acl", format, v.DeviceID, v.Unchanged, v.Updated, v.Added, v.Deleted, v.Failed, v.Errored+len(warnings))
		}

		for k, v := range report {
			for _, err := range v.Errors {
				errorf("load-acl", "%v  %v", k, err)
			}
		}
	}

	return nil
}

func (cmd *LoadACL) load(u uhppote.IUHPPOTE, acl lib.ACL) (map[uint32]lib.Report, []error) {
	f := func(u uhppote.IUHPPOTE, list lib.ACL) (map[uint32]lib.Report, []error) {
		if cmd.withPIN {
			return lib.PutACLWithPIN(u, list, false)
		} else {
			return lib.PutACL(u, list, false)
		}
	}

	return f(u, acl)
}

func report2audit(report map[uint32]lib.Report) []db.AuditRecord {
	now := time.Now()
	recordset := []db.AuditRecord{}

	auditRecord := func(controller uint32, card uint32, status string) db.AuditRecord {
		return db.AuditRecord{
			Timestamp:  now,
			Operation:  "load",
			Controller: controller,
			CardNumber: card,
			Status:     status,
			Card:       "",
		}
	}

	for controller, v := range report {
		for _, card := range v.Updated {
			recordset = append(recordset, auditRecord(controller, card, "updated"))
		}

		for _, card := range v.Added {
			recordset = append(recordset, auditRecord(controller, card, "added"))
		}

		for _, card := range v.Deleted {
			recordset = append(recordset, auditRecord(controller, card, "deleted"))
		}

		for _, card := range v.Failed {
			recordset = append(recordset, auditRecord(controller, card, "failed"))
		}

		for _, card := range v.Errored {
			recordset = append(recordset, auditRecord(controller, card, "error"))
		}
	}

	return recordset
}

func report2log(report map[uint32]lib.Report) []db.LogRecord {
	now := time.Now()
	recordset := []db.LogRecord{}

	logRecord := func(controller uint32, row lib.Report) db.LogRecord {
		detail := fmt.Sprintf("unchanged:%-4v updated:%-4v   added:%-4v   deleted:%-4v failed:%-4v errors:%-4v",
			len(row.Unchanged),
			len(row.Updated),
			len(row.Added),
			len(row.Deleted),
			len(row.Failed),
			len(row.Errored))

		return db.LogRecord{
			Timestamp:  now,
			Operation:  "load",
			Controller: controller,
			Detail:     detail,
		}
	}

	for controller, v := range report {
		recordset = append(recordset, logRecord(controller, v))
	}

	return recordset
}
