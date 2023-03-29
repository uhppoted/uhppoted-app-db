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
	return "--dsn <DSN> --with-pin --file <file>"
}

func (cmd *GetACL) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s [--debug] [--config <file>] get-acl [--dsn <DSN>] [--with-pin] [--file <file>]\n", APP)
	fmt.Println()
	fmt.Println("  Retrieves an access control list from a database and optionally saves the ACL to a TSV file")
	fmt.Println()

	helpOptions(cmd.FlagSet())

	fmt.Println()
	fmt.Println("  Examples:")
	fmt.Println(`    uhppote-app-db --debug get-acl --dsn "sqlite3:" --file "ACL.tsv"`)
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
		if t, err := sqlite3.GetACL(cmd.dsn[8:]); err != nil {
			return err
		} else if t == nil {
			return fmt.Errorf("Invalid ACL table (%v)", table)
		} else {
			table = t
		}
	}

	f := func(table *lib.Table, devices []uhppote.Device) (*lib.ACL, []error, error) {
		if cmd.withPIN {
			return lib.ParseTable(table, devices, false)
		} else {
			return lib.ParseTable(table, devices, false)
		}
	}

	if list, warnings, err := f(table, devices); err != nil {
		return err
	} else if list == nil {
		return fmt.Errorf("error creating ACL from DB table (%v)", list)
	} else {
		for _, w := range warnings {
			warnf("%v", w.Error())
		}

		fmt.Printf(">>>>>>> ACL: %v\n", list)
		return nil
	}

	// if cmd.debug {
	//     if cmd.withPIN {
	//         fmt.Printf("ACL:\n%s\n", string(ACL.AsTableWithPIN().MarshalTextIndent("  ", " ")))
	//     } else {
	//         fmt.Printf("ACL:\n%s\n", string(ACL.AsTable().MarshalTextIndent("  ", " ")))
	//     }
	// }

	// for _, w := range warnings {
	//     warnf("%v", w.Error())
	// }

	// // ... write to stdout
	// if cmd.file == "" {
	//     fmt.Fprintln(os.Stdout, string(asTable(ACL).MarshalTextIndent("  ", " ")))
	//     return nil
	// }

	// // ... write to TSV file
	// asTSV := func(a *acl.ACL, w io.Writer) error {
	//     if cmd.withPIN {
	//         return a.ToTSVWithPIN(w)
	//     } else {
	//         return a.ToTSV(w)
	//     }
	// }

	// var b bytes.Buffer
	// if err := asTSV(ACL, &b); err != nil {
	//     return fmt.Errorf("error creating TSV file (%v)", err)
	// }

	// if err := write(cmd.file, b.Bytes()); err != nil {
	//     return err
	// }

	// infof("ACL saved to %s", cmd.file)

	return nil
}
