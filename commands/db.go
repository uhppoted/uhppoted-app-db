package commands

import (
	"fmt"
	"strings"

	core "github.com/uhppoted/uhppote-core/types"
	lib "github.com/uhppoted/uhppoted-lib/acl"

	"github.com/uhppoted/uhppoted-app-db/db"
	"github.com/uhppoted/uhppoted-app-db/db/mssql"
	"github.com/uhppoted/uhppoted-app-db/db/mysql"
	"github.com/uhppoted/uhppoted-app-db/db/sqlite3"
)

func fromDSN(dsn string) (db.DB, error) {
	switch {
	case strings.HasPrefix(dsn, "sqlite3://"):
		return sqlite3.NewDB(dsn[10:]), nil

	case strings.HasPrefix(dsn, "sqlserver://"):
		return mssql.NewDB(dsn), nil

	case strings.HasPrefix(dsn, "mysql://"):
		return mysql.NewDB(dsn[8:]), nil

	default:
		return nil, fmt.Errorf("unsupported DSN (%v)", dsn)
	}
}

func getACL(dsn string, table string, withPIN bool) (lib.Table, error) {
	if dbi, err := fromDSN(dsn); err != nil {
		return lib.Table{}, err
	} else if t, err := dbi.GetACL(table, withPIN); err != nil {
		return lib.Table{}, err
	} else if t == nil {
		return lib.Table{}, fmt.Errorf("invalid ACL table (%v)", t)
	} else {
		return *t, nil
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

func getEvents(dsn string, table string, controller uint32) ([]uint32, error) {
	switch {
	case strings.HasPrefix(dsn, "sqlite3://"):
		if events, err := sqlite3.GetEvents(dsn[10:], table, controller); err != nil {
			return nil, err
		} else {
			return events, nil
		}

	case strings.HasPrefix(dsn, "sqlserver://"):
		if events, err := mssql.GetEvents(dsn, table, controller); err != nil {
			return nil, err
		} else {
			return events, nil
		}

	default:
		return nil, fmt.Errorf("unsupported DSN (%v)", dsn)
	}
}

func putEvents(dsn string, table string, events []core.Event) error {
	switch {
	case strings.HasPrefix(dsn, "sqlite3://"):
		if N, err := sqlite3.StoreEvents(dsn[10:], table, events); err != nil {
			return err
		} else if N == 1 {
			infof("get-events", "Stored %v event to DB events table", N)
		} else {
			infof("get-events", "Stored %v events to DB events table", N)
		}

	case strings.HasPrefix(dsn, "sqlserver://"):
		if N, err := mssql.StoreEvents(dsn, table, events); err != nil {
			return err
		} else if N == 1 {
			infof("put-acl", "Stored %v event to DB events table", N)
		} else {
			infof("put-acl", "Stored %v events to DB events table", N)
		}

	default:
		return fmt.Errorf("unsupported DSN (%v)", dsn)
	}

	return nil
}

func stashToAudit(dsn string, table string, recordset []db.AuditRecord) error {
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

func stashToLog(dsn string, table string, recordset []db.LogRecord) error {
	switch {
	case strings.HasPrefix(dsn, "sqlite3://"):
		if N, err := sqlite3.Log(dsn[10:], table, recordset); err != nil {
			return err
		} else if N == 1 {
			infof("log", "Added 1 record to operations log")
		} else {
			infof("log", "Added %v records to operations log", N)
		}

	case strings.HasPrefix(dsn, "sqlserver://"):
		if N, err := mssql.Log(dsn, table, recordset); err != nil {
			return err
		} else if N == 1 {
			infof("log", "Added 1 record to operations log")
		} else {
			infof("log", "Added %v records to operations log", N)
		}

	default:
		return fmt.Errorf("unsupported DSN (%v)", dsn)
	}

	return nil
}
