package commands

import (
	"fmt"
	"strings"

	lib "github.com/uhppoted/uhppoted-lib/acl"

	"github.com/uhppoted/uhppoted-app-db/db/sqlite3"
)

func getACL(dsn string, withPIN bool) (lib.Table, error) {
	switch {
	case strings.HasPrefix(dsn, "sqlite3:"):
		if table, err := sqlite3.GetACL(dsn[8:], withPIN); err != nil {
			return lib.Table{}, err
		} else if table == nil {
			return lib.Table{}, fmt.Errorf("invalid ACL table (%v)", table)
		} else {
			return *table, nil
		}

	default:
		return lib.Table{}, fmt.Errorf("unsupported DSN (%v)", dsn)
	}
}

func putACL(dsn string, acl lib.Table, withPIN bool) error {
	switch {
	case strings.HasPrefix(dsn, "sqlite3:"):
		if N, err := sqlite3.PutACL(dsn[8:], acl, withPIN); err != nil {
			return err
		} else if N == 1 {
			infof("put-acl", "Stored %v card to DB ACL table", N)
		} else {
			infof("put-acl", "Stored %v cards to DB ACL table", N)
		}

	default:
		return fmt.Errorf("unsupported DSN (%v)", dsn)
	}

	return nil
}
