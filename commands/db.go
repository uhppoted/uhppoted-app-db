package commands

import (
	"fmt"
	"strings"

	lib "github.com/uhppoted/uhppoted-lib/acl"

	"github.com/uhppoted/uhppoted-app-db/db"
	"github.com/uhppoted/uhppoted-app-db/db/mssql"
	"github.com/uhppoted/uhppoted-app-db/db/sqlite3"
)

func getACL(dsn string, table string, withPIN bool) (lib.Table, error) {
	switch {
	case strings.HasPrefix(dsn, "sqlite3://"):
		if t, err := sqlite3.GetACL(dsn[10:], table, withPIN); err != nil {
			return lib.Table{}, err
		} else if t == nil {
			return lib.Table{}, fmt.Errorf("invalid ACL table (%v)", t)
		} else {
			return *t, nil
		}

	case strings.HasPrefix(dsn, "sqlserver://"):
		if t, err := mssql.GetACL(dsn, table, withPIN); err != nil {
			return lib.Table{}, err
		} else if t == nil {
			return lib.Table{}, fmt.Errorf("invalid ACL table (%v)", t)
		} else {
			return *t, nil
		}

	default:
		return lib.Table{}, fmt.Errorf("unsupported DSN (%v)", dsn)
	}
}

func putACL(dsn string, table string, acl lib.Table, withPIN bool) error {
	switch {
	case strings.HasPrefix(dsn, "sqlite3://"):
		if N, err := sqlite3.PutACL(dsn[10:], table, acl, withPIN); err != nil {
			return err
		} else if N == 1 {
			infof("put-acl", "Stored %v card to DB ACL table", N)
		} else {
			infof("put-acl", "Stored %v cards to DB ACL table", N)
		}

	case strings.HasPrefix(dsn, "sqlserver://"):
		if N, err := mssql.PutACL(dsn, table, acl, withPIN); err != nil {
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

func stash(operation string, dsn string, table string, recordset []db.AuditRecord) error {
	switch {
	case strings.HasPrefix(dsn, "sqlite3://"):
		if N, err := sqlite3.AuditTrail(dsn[10:], table, recordset); err != nil {
			return err
		} else if N == 1 {
			infof("audit", "Added 1 record to audit trail")
		} else {
			infof("audit", "Added %v records to audit trail", N)
		}

	case strings.HasPrefix(dsn, "sqlserver://"):
		if N, err := mssql.AuditTrail(dsn, table, recordset); err != nil {
			return err
		} else if N == 1 {
			infof("audit", "Added 1 record to audit trail")
		} else {
			infof("audit", "Added %v records to audit trail", N)
		}

	default:
		return fmt.Errorf("unsupported DSN (%v)", dsn)
	}

	return nil
}
