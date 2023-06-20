package commands

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/uhppoted/uhppote-core/uhppote"
	lib "github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/config"
)

var GetACLCmd = GetACL{
	command: command{
		name:        "get-acl",
		description: "Retrieves an access control list from a database and (optionally) saves it to a file",
		usage:       "[--with-pin] --dsn <DSN> [--table:ACL <table>] [--file <file>]",

		dsn:      "",
		tableACL: "ACL",
		withPIN:  false,
		lockfile: "",
		config:   config.DefaultConfig,
		debug:    false,
	},
}

type GetACL struct {
	command
	file string
}

func (cmd *GetACL) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s [--debug] [--config <file>] get-acl [--with-pin] --dsn <DSN> [--table:ACL <table>] [--file <file>]\n", APP)
	fmt.Println()
	fmt.Println("  Retrieves an access control list from a database and optionally saves the ACL to a TSV file")
	fmt.Println()

	helpOptions(cmd.FlagSet())

	fmt.Println()
	fmt.Println("  Examples:")
	fmt.Println(`    uhppote-app-db --debug get-acl --with-pin --dsn "sqlite3://./db/ACL.db" --file "ACL.tsv"`)
	fmt.Println(`    uhppote-app-db --debug get-acl --with-pin --dsn "sqlite3://./db/ACL.db" --table:ACL ACL --file "ACL.tsv"`)
	fmt.Println()
}

func (cmd *GetACL) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("get-acl", flag.ExitOnError)

	flagset.StringVar(&cmd.dsn, "dsn", cmd.dsn, "DSN for database")
	flagset.StringVar(&cmd.tableACL, "table:ACL", cmd.tableACL, "ACL table name. Defaults to ACL")
	flagset.StringVar(&cmd.file, "file", cmd.file, "Optional TSV filepath. Defaults to stdout")
	flagset.BoolVar(&cmd.withPIN, "with-pin", cmd.withPIN, "Include card keypad PIN code in retrieved ACL information")
	flagset.StringVar(&cmd.lockfile, "lockfile", cmd.lockfile, "Filepath for lock file. Defaults to <tmp>/uhppoted-app-db.lock")

	return flagset
}

func (cmd *GetACL) Execute(args ...any) error {
	options := args[0].(*Options)

	cmd.config = options.Config
	cmd.debug = options.Debug

	// ... check parameters
	if strings.TrimSpace(cmd.dsn) == "" {
		return fmt.Errorf("invalid database DSN")
	}

	if strings.TrimSpace(cmd.tableACL) == "" {
		return fmt.Errorf("invalid ACL table")
	}

	// ... locked?
	if kraken, err := lock(cmd.lockfile); err != nil {
		return err
	} else {
		defer func() {
			infof("get-acl", "Removing lockfile")
			kraken.Release()
		}()
	}

	// ... get config
	conf := config.NewConfig()
	if err := conf.Load(cmd.config); err != nil {
		return fmt.Errorf("could not load configuration (%v)", err)
	}

	_, devices := getDevices(conf, cmd.debug)

	// ... retrieve ACL from DB
	f := func(table lib.Table, devices []uhppote.Device) (*lib.ACL, []error, error) {
		if cmd.withPIN {
			return lib.ParseTable(&table, devices, false)
		} else {
			return lib.ParseTable(&table, devices, false)
		}
	}

	if table, err := getACL(cmd.dsn, cmd.tableACL, cmd.withPIN); err != nil {
		return err
	} else if acl, warnings, err := f(table, devices); err != nil {
		return err
	} else if acl == nil {
		return fmt.Errorf("error creating ACL from DB table (%v)", acl)
	} else {
		for _, w := range warnings {
			warnf("get-acl", "%v", w.Error())
		}

		// ... write to TSV file
		if cmd.file != "" {
			var b bytes.Buffer

			if err := table.ToTSV(&b); err != nil {
				return fmt.Errorf("error creating TSV file (%v)", err)
			} else if err := write(cmd.file, b.Bytes()); err != nil {
				return err
			}

			infof("get-acl", "ACL saved to %v", cmd.file)
			return nil
		}

		// ... write to stdout
		fmt.Fprintln(os.Stdout, string(table.MarshalTextIndent("  ", " ")))

		if cmd.debug {
			acl.Print(os.Stdout)
		}
	}

	return nil
}
