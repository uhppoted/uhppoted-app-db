package sqlite3

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func GetACL(dsn string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if _, err := os.Stat(dsn); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("sqlite3 database %v does not exist", dsn)
	} else if err != nil {
		return err
	}

	if dbc, err := open(dsn, MaxLifetime, MaxIdle, MaxOpen); err != nil {
		return err
	} else if dbc == nil {
		return fmt.Errorf("invalid sqlite3 DB pool (%v)", dbc)
	} else if prepared, err := dbc.Prepare(sqlAclGet); err != nil {
		return err
	} else if rs, err := prepared.QueryContext(ctx); err != nil {
		return err
	} else if rs == nil {
		return fmt.Errorf("invalid resultset (%v)", rs)
	} else {
		defer rs.Close()

		if columns, err := rs.Columns(); err != nil {
			return err
		} else if types, err := rs.ColumnTypes(); err != nil {
			return err
		} else {
			for rs.Next() {
				if record, err := row2record(rs, columns, types); err != nil {
					return err
				} else if record == nil {
					return fmt.Errorf("invalid record (%v)", record)
				} else {
					fmt.Printf(">>>>>> %v\n", record)
				}
			}
		}
	}

	return nil
}
