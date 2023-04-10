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
)

var StoreACLCmd = StoreACL{
	command: command{
		name:        "store-acl",
		description: "Retrieves the ACL from a set of access controllers and stores it in a database table",
		usage:       "--with-pin --dsn <DSN>",

		dsn:      "",
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
	fmt.Printf("  Usage: %s [--debug] [--config <file>] store-acl [--with-pin] --dsn <DSN>\n", APP)
	fmt.Println()
	fmt.Println("  Retrieves the ACL from a set of access controllers and stores it in a database table")
	fmt.Println()

	helpOptions(cmd.FlagSet())

	fmt.Println()
	fmt.Println("  Examples:")
	fmt.Println(`    uhppote-app-db --debug store-acl --with-pin --dsn "sqlite3:./db/ACL.db::ACL"`)
	fmt.Println()
}

func (cmd *StoreACL) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("store-acl", flag.ExitOnError)

	flagset.StringVar(&cmd.dsn, "dsn", cmd.dsn, "DSN for database")
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
			infof("store-acl", "Removing lockfile '%v'", lockFile.File)
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
	} else if err := putACL(cmd.dsn, *acl, cmd.withPIN); err != nil {
		return err
	} else {
		infof("store-acl", "Updated DB ACL table")
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
