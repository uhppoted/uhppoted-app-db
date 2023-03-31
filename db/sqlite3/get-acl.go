package sqlite3

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"

	lib "github.com/uhppoted/uhppoted-lib/acl"
)

func GetACL(dsn string, withPIN bool) (*lib.Table, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if _, err := os.Stat(dsn); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("sqlite3 database %v does not exist", dsn)
	} else if err != nil {
		return nil, err
	}

	if dbc, err := open(dsn, MaxLifetime, MaxIdle, MaxOpen); err != nil {
		return nil, err
	} else if dbc == nil {
		return nil, fmt.Errorf("invalid sqlite3 DB pool (%v)", dbc)
	} else if prepared, err := dbc.Prepare(sqlAclGet); err != nil {
		return nil, err
	} else if rs, err := prepared.QueryContext(ctx); err != nil {
		return nil, err
	} else if rs == nil {
		return nil, fmt.Errorf("invalid resultset (%v)", rs)
	} else {
		defer rs.Close()

		if columns, err := rs.Columns(); err != nil {
			return nil, err
		} else if types, err := rs.ColumnTypes(); err != nil {
			return nil, err
		} else {
			recordset := []record{}

			for rs.Next() {
				if record, err := row2record(rs, columns, types); err != nil {
					return nil, err
				} else if record == nil {
					return nil, fmt.Errorf("invalid record (%v)", record)
				} else {
					recordset = append(recordset, record)
				}
			}

			return makeTable(columns, recordset, withPIN)
		}
	}
}
