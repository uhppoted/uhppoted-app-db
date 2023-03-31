package commands

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/uhppoted/uhppote-core/uhppote"
	lib "github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/config"
	"github.com/uhppoted/uhppoted-lib/lockfile"

	"github.com/uhppoted/uhppoted-app-db/db/sqlite3"
)

var GetACLCmd = GetACL{
	config:   config.DefaultConfig,
	dsn:      "",
	withPIN:  false,
	lockfile: "",
	debug:    false,
}

type GetACL struct {
	config   string
	dsn      string
	file     string
	withPIN  bool
	lockfile string
	debug    bool
}

func (cmd *GetACL) Name() string {
	return "get-acl"
}

func (cmd *GetACL) Description() string {
	return "Retrieves an access control list from a database and (optionally) saves it to a file"
}

func (cmd *GetACL) Usage() string {
	return "--with-pin --dsn <DSN> --file <file>"
}

func (cmd *GetACL) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s [--debug] [--config <file>] get-acl [--with-pin] --dsn <DSN> [--file <file>]\n", APP)
	fmt.Println()
	fmt.Println("  Retrieves an access control list from a database and optionally saves the ACL to a TSV file")
	fmt.Println()

	helpOptions(cmd.FlagSet())

	fmt.Println()
	fmt.Println("  Examples:")
	fmt.Println(`    uhppote-app-db --debug get-acl --with-pin --dsn "sqlite3:./db/ACL.db" --file "ACL.tsv"`)
	fmt.Println()
}

func (cmd *GetACL) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("get-acl", flag.ExitOnError)

	flagset.StringVar(&cmd.dsn, "dsn", cmd.dsn, "DSN for database")
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

	// ... locked?
	lockFile := config.Lockfile{
		File:   filepath.Join(os.TempDir(), "uhppoted-app-db.lock"),
		Remove: lockfile.RemoveLockfile,
	}

	if cmd.lockfile != "" {
		lockFile = config.Lockfile{
			File:   cmd.lockfile,
			Remove: lockfile.RemoveLockfile,
		}
	}

	if kraken, err := lockfile.MakeLockFile(lockFile); err != nil {
		return err
	} else {
		defer func() {
			infof("get-acl", "Removing lockfile '%v'", lockFile.File)
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
	var table *lib.Table

	switch {
	case strings.HasPrefix(cmd.dsn, "sqlite3:"):
		if t, err := sqlite3.GetACL(cmd.dsn[8:], cmd.withPIN); err != nil {
			return err
		} else if t == nil {
			return fmt.Errorf("invalid ACL table (%v)", table)
		} else {
			table = t
		}

	default:
		return fmt.Errorf("unsupported DSN (%v)", cmd.dsn)
	}

	f := func(table *lib.Table, devices []uhppote.Device) (*lib.ACL, []error, error) {
		if cmd.withPIN {
			return lib.ParseTable(table, devices, false)
		} else {
			return lib.ParseTable(table, devices, false)
		}
	}

	if acl, warnings, err := f(table, devices); err != nil {
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
