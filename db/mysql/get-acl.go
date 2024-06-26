package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	lib "github.com/uhppoted/uhppoted-lib/acl"
)

func GetACL(dsn string, table string, withPIN bool) (*lib.Table, error) {
	if dbc, err := open(dsn, MaxLifetime, MaxIdle, MaxOpen); err != nil {
		return nil, err
	} else if dbc == nil {
		return nil, fmt.Errorf("invalid %v DB (%v)", "MySQL", dbc)
	} else {
		return get(dbc, table, withPIN)
	}
}

func get(dbc *sql.DB, table string, withPIN bool) (*lib.Table, error) {
	sql := fmt.Sprintf(`SELECT * FROM %v;`, table)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if prepared, err := dbc.Prepare(sql); err != nil {
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
