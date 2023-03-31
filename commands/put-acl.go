package commands

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-app-db/db/sqlite3"
	lib "github.com/uhppoted/uhppoted-lib/acl"
	"github.com/uhppoted/uhppoted-lib/config"
	"github.com/uhppoted/uhppoted-lib/lockfile"
)

var PutACLCmd = PutACL{
	config:   config.DefaultConfig,
	dsn:      "",
	file:     "",
	withPIN:  false,
	lockfile: "",
	// strict:   false,
	debug: false,
}

type PutACL struct {
	config   string
	dsn      string
	file     string
	withPIN  bool
	lockfile string
	// strict   bool
	debug bool
}

func (cmd *PutACL) Name() string {
	return "put-acl"
}

func (cmd *PutACL) Description() string {
	return "Stores an access control list in a TSV file to a database"
}

func (cmd *PutACL) Usage() string {
	return "--with-pin --dsn <DSN> --file <file>"
}

func (cmd *PutACL) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s [--debug] [--config <file>] put-acl [--with-pin] [--strict] --file <file> --dsn <DSN>\n", APP)
	fmt.Println()
	fmt.Println("  Stores an access control list in a TSV file to a database")
	fmt.Println()

	helpOptions(cmd.FlagSet())

	fmt.Println()
	fmt.Println("  Examples:")
	fmt.Println(`    uhppote-app-db --debug put-acl --with-pin --file "ACL.tsv" --dsn "sqlite3:./db/ACL.db"`)
	fmt.Println()
}

func (cmd *PutACL) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("put-acl", flag.ExitOnError)

	flagset.StringVar(&cmd.dsn, "dsn", cmd.dsn, "DSN for database")
	flagset.StringVar(&cmd.file, "file", cmd.file, "Optional TSV filepath. Defaults to stdout")
	flagset.BoolVar(&cmd.withPIN, "with-pin", cmd.withPIN, "Include card keypad PIN code in retrieved ACL information")
	// flagset.BoolVar(&cmd.withPIN, "strict", cmd.strict, "Fails if the TSV file includes duplicate card numbers.")
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
			infof("put-acl", "Removing lockfile '%v'", lockFile.File)
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

		switch {
		case strings.HasPrefix(cmd.dsn, "sqlite3:"):
			if err := sqlite3.PutACL(cmd.dsn[8:], acl, cmd.withPIN); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported DSN (%v)", cmd.dsn)
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
			Records: records,
		}, nil, nil
	}
}
