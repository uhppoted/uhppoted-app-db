package commands

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-app-db/db"
	lib "github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/config"
)

var StoreACLCmd = StoreACL{
	command: command{
		name:        "store-acl",
		description: "Retrieves the ACL from a set of access controllers and stores it in a database table",
		usage:       "--with-pin --dsn <DSN> [--table:ACL <table>] [--table:log <table>]",

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
}

type StoreACL struct {
	command
}

func (cmd *StoreACL) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s [--debug] [--config <file>] store-acl [--with-pin] --dsn <DSN> [--table:ACL <table>] [--table:log <table>]\n", APP)
	fmt.Println()
	fmt.Println("  Retrieves the ACL from a set of access controllers and stores it in a database table")
	fmt.Println()

	helpOptions(cmd.FlagSet())

	fmt.Println()
	fmt.Println("  Examples:")
	fmt.Println(`    uhppote-app-db --debug store-acl --with-pin --dsn "sqlite3://./db/ACL.db::ACL"`)
	fmt.Println(`    uhppote-app-db --debug store-acl --with-pin --dsn "sqlite3://./db/ACL.db::ACL" --table:ACL ACL2 --table:log OpsLog`)
	fmt.Println()
}

func (cmd *StoreACL) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("store-acl", flag.ExitOnError)

	flagset.StringVar(&cmd.dsn, "dsn", cmd.dsn, "DSN for database")
	flagset.StringVar(&cmd.tables.ACL, "table:ACL", cmd.tables.ACL, "ACL table name. Defaults to ACL")
	flagset.StringVar(&cmd.tables.Log, "table:log", cmd.tables.Log, "Operations log table name. Defaults to ''")
	flagset.BoolVar(&cmd.withPIN, "with-pin", cmd.withPIN, "Include card keypad PIN code in retrieved ACL information")
	flagset.StringVar(&cmd.lockfile, "lockfile", cmd.lockfile, "Filepath for lock file. Defaults to <tmp>/uhppoted-app-db.lock")

	return flagset
}

func (cmd *StoreACL) Execute(args ...any) error {
	options := args[0].(*Options)

	cmd.config = options.Config
	cmd.debug = options.Debug

	// ... check parameters
	if strings.TrimSpace(cmd.dsn) == "" {
		return fmt.Errorf("missing database DSN")
	}

	// ... locked?
	if kraken, err := lock(cmd.lockfile); err != nil {
		return err
	} else {
		defer func() {
			infof("store-acl", "removing lockfile")
			kraken.Release()
		}()
	}

	// ... get config
	conf := config.NewConfig()
	if err := conf.Load(cmd.config); err != nil {
		return fmt.Errorf("could not load configuration (%v)", err)
	}

	u, devices := getDevices(conf, cmd.debug)

	// ... retrieve ACL from controllers
	if acl, err := cmd.getACL(u, devices); err != nil {
		return err
	} else if acl == nil {
		return fmt.Errorf("invalid ACL (%v)", acl)
	} else if err := putACL(cmd.dsn, cmd.tables.ACL, *acl, cmd.withPIN); err != nil {
		return err
	} else {
		infof("store-acl", "Updated DB ACL table")

		if cmd.tables.Log != "" {
			recordset := []db.LogRecord{
				db.LogRecord{
					Timestamp: time.Now(),
					Operation: "store-acl",
					Detail:    fmt.Sprintf("records:%v", len(acl.Records)),
				},
			}

			if err := stashToLog(cmd.dsn, cmd.tables.Log, recordset); err != nil {
				return err
			}
		}
	}

	return nil
}

func (cmd *StoreACL) getACL(u uhppote.IUHPPOTE, devices []uhppote.Device) (*lib.Table, error) {
	acl, errors := lib.GetACL(u, devices)
	if len(errors) > 0 {
		return nil, fmt.Errorf("%v", errors)
	}

	for k, l := range acl {
		infof("store-acl", "%v  Retrieved %v records", k, len(l))
	}

	if cmd.withPIN {
		return lib.MakeTableWithPIN(acl, devices)
	} else {
		return lib.MakeTable(acl, devices)
	}
}
