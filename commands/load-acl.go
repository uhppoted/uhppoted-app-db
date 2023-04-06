package commands

import (
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

var LoadACLCmd = LoadACL{
	config:   config.DefaultConfig,
	dsn:      "",
	withPIN:  false,
	lockfile: "",
	debug:    false,
}

type LoadACL struct {
	config   string
	dsn      string
	withPIN  bool
	lockfile string
	debug    bool
}

func (cmd *LoadACL) Name() string {
	return "load-acl"
}

func (cmd *LoadACL) Description() string {
	return "Retrieves an access control list from a database and updates the configured set of access controllers"
}

func (cmd *LoadACL) Usage() string {
	return "--with-pin --dsn <DSN>"
}

func (cmd *LoadACL) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s [--debug] [--config <file>] load-acl [--with-pin] --dsn <DSN>\n", APP)
	fmt.Println()
	fmt.Println("  Retrieves an access control list from a database and updates the configured set of access controllers")
	fmt.Println()

	helpOptions(cmd.FlagSet())

	fmt.Println()
	fmt.Println("  Examples:")
	fmt.Println(`    uhppote-app-db --debug load-acl --with-pin --dsn "sqlite3:./db/ACL.db"`)
	fmt.Println()
}

func (cmd *LoadACL) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("load-acl", flag.ExitOnError)

	flagset.StringVar(&cmd.dsn, "dsn", cmd.dsn, "DSN for database")
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
			infof("load-acl", "Removing lockfile '%v'", lockFile.File)
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
