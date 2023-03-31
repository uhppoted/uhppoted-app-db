package sqlite3

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	lib "github.com/uhppoted/uhppoted-lib/acl"
)

func PutACL(dsn string, table lib.Table, withPIN bool) error {
	// ... format SQL
	columns := make([]string, len(table.Header))
	values := make([]string, len(table.Header))
	conflicts := []string{}

	for i, h := range table.Header {
		col := strings.ReplaceAll(h, " ", "")
		columns[i] = col
		values[i] = "?"

		if strings.ToLower(col) != "cardnumber" {
			conflicts = append(conflicts, fmt.Sprintf("%v=excluded.%v", col, col))
		}
	}

	insert := fmt.Sprintf("INSERT INTO ACLx (%v) VALUES (%v) ON CONFLICT(CardNumber) DO UPDATE SET %v;",
		strings.Join(columns, ","),
		strings.Join(values, ","),
		strings.Join(conflicts, ","))

	// ... execute
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
		return fmt.Errorf("invalid sqlite3 DB (%v)", dbc)
	} else if prepared, err := dbc.Prepare(insert); err != nil {
		return err
	} else {
		for _, row := range table.Records {
			record := make([]any, len(row))
			for i, v := range row {
				record[i] = v
			}

			if _, err := prepared.ExecContext(ctx, record...); err != nil {
				return err
			}
		}
	}

	return nil
}
