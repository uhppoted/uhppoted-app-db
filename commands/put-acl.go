package commands

import (
	"encoding/csv"
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

var PutACLCmd = PutACL{
	command: command{
		name:        "put-acl",
		description: "Stores an access control list in a TSV file to a database",
		usage:       "[--with-pin] --dsn <DSN> [--table:ACL <table>] [--table:log <table>] [--file <file>]",

		dsn: "",
		tables: tables{
			ACL: "ACL",
			Log: "",
		},
		withPIN:  false,
		lockfile: "",
		config:   config.DefaultConfig,
		debug:    false,
	},

	file: "",
}

type PutACL struct {
	command
	file string
}

func (cmd *PutACL) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s [--debug] [--config <file>] put-acl [--with-pin] --file <file> --dsn <DSN> [--table:ACL <table>] [--table:log <table>]\n", APP)
	fmt.Println()
	fmt.Println("  Stores an access control list in a TSV file to a database")
	fmt.Println()

	helpOptions(cmd.FlagSet())

	fmt.Println()
	fmt.Println("  Examples:")
	fmt.Println(`    uhppote-app-db --debug put-acl --with-pin --file "ACL.tsv" --dsn "sqlite3://./db/ACL.db"`)
	fmt.Println(`    uhppote-app-db --debug put-acl --with-pin --file "ACL.tsv" --dsn "sqlite3://./db/ACL.db" --table:ACL ACL2  --table:Log OpsLog`)
	fmt.Println()
}

func (cmd *PutACL) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("put-acl", flag.ExitOnError)

	flagset.StringVar(&cmd.dsn, "dsn", cmd.dsn, "DSN for database")
	flagset.StringVar(&cmd.tables.ACL, "table:ACL", cmd.tables.ACL, "ACL table name. Defaults to ACL")
	flagset.StringVar(&cmd.tables.Log, "table:log", cmd.tables.Log, "Operations log table name. Defaults to ''")
	flagset.StringVar(&cmd.file, "file", cmd.file, "Optional TSV filepath. Defaults to stdout")
	flagset.BoolVar(&cmd.withPIN, "with-pin", cmd.withPIN, "Include card keypad PIN code in retrieved ACL information")
	flagset.StringVar(&cmd.lockfile, "lockfile", cmd.lockfile, "Filepath for lock file. Defaults to <tmp>/uhppoted-app-db.lock")

	return flagset
}

func (cmd *PutACL) Execute(args ...any) error {
	options := args[0].(*Options)

	cmd.config = options.Config
	cmd.debug = options.Debug

	// ... check parameters
	if strings.TrimSpace(cmd.file) == "" {
		return fmt.Errorf("missing TSV file")
	}

	if strings.TrimSpace(cmd.dsn) == "" {
		return fmt.Errorf("missing database DSN")
	}

	if strings.TrimSpace(cmd.tables.ACL) == "" {
		return fmt.Errorf("missing ACL table")
	}

	// ... locked?
	if kraken, err := lock(cmd.lockfile); err != nil {
		return err
	} else {
		defer func() {
			infof("put-acl", "removing lockfile")
			kraken.Release()
		}()
	}

	// ... get config
	conf := config.NewConfig()
	if err := conf.Load(cmd.config); err != nil {
		return fmt.Errorf("could not load configuration (%v)", err)
	}

	_, devices := getDevices(conf, cmd.debug)

	// ... retrieve ACL from TSV file
	if acl, warnings, err := cmd.getACL(devices); err != nil {
		return err
	} else {
		for _, w := range warnings {
			warnf("put-acl", "%v", w.Error())
		}

		if err := putACL(cmd.dsn, cmd.tables.ACL, acl, cmd.withPIN); err != nil {
			return err
		} else {
			infof("put-acl", "Updated DB ACL table from %v", cmd.file)

			if cmd.tables.Log != "" {
				recordset := []db.LogRecord{
					db.LogRecord{
						Timestamp: time.Now(),
						Operation: "put-acl",
						Detail:    fmt.Sprintf("records:%v", len(acl.Records)),
					},
				}

				if err := stashToLog(cmd.dsn, cmd.tables.Log, recordset); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (cmd *PutACL) getACL(devices []uhppote.Device) (lib.Table, []error, error) {
	if f, err := os.Open(cmd.file); err != nil {
		return lib.Table{}, nil, err
	} else {
		defer f.Close()

		r := csv.NewReader(f)
		r.Comma = '\t'

		records, err := r.ReadAll()
		if err != nil {
			return lib.Table{}, nil, err
		} else if len(records) == 0 {
			return lib.Table{}, nil, fmt.Errorf("TSV file is empty")
		} else if len(records) < 1 {
			return lib.Table{}, nil, fmt.Errorf("TSV file missing header")
		}

		// ... header
		header := make([]string, len(records[0]))

		for i, v := range records[0] {
			header[i] = fmt.Sprintf("%v", v)
		}

		// ... records
		rows := make([][]string, 0)

		for _, record := range records[1:] {
			row := make([]string, len(record))

			for i, v := range record {
				row[i] = fmt.Sprintf("%v", v)
			}

			rows = append(rows, row)
		}

		return lib.Table{
			Header:  header,
			Records: rows,
		}, nil, nil
	}
}
