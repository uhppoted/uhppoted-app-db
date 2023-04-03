package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	lib "github.com/uhppoted/uhppoted-lib/acl"
)

func PutACL(dsn string, table lib.Table, withPIN bool) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if _, err := os.Stat(dsn); errors.Is(err, os.ErrNotExist) {
		return 0, fmt.Errorf("sqlite3 database %v does not exist", dsn)
	} else if err != nil {
		return 0, err
	}

	if dbc, err := open(dsn, MaxLifetime, MaxIdle, MaxOpen); err != nil {
		return 0, err
	} else if dbc == nil {
		return 0, fmt.Errorf("invalid sqlite3 DB (%v)", dbc)
	} else if tx, err := dbc.BeginTx(ctx, nil); err != nil {
		return 0, err
	} else if _, err := clear(dbc, tx, "ACLx"); err != nil {
		return 0, err
	} else if count, err := insert(dbc, tx, "ACLx", table); err != nil {
		return 0, err
	} else if err := tx.Commit(); err != nil {
		return 0, err
	} else {
		return count, nil
	}
}

func clear(dbc *sql.DB, tx *sql.Tx, table string) (int64, error) {
	sql := fmt.Sprintf("DELETE FROM %v;", table)

	if prepared, err := dbc.Prepare(sql); err != nil {
		return 0, err
	} else if result, err := tx.Stmt(prepared).Exec(); err != nil {
		return 0, err
	} else if N, err := result.RowsAffected(); err != nil {
		return N, err
	} else {
		return N, nil
	}
}

func insert(dbc *sql.DB, tx *sql.Tx, table string, recordset lib.Table) (int, error) {
	columns := []string{"CardNumber", "StartDate", "EndDate"}
	index := map[string]int{}

	for i, h := range recordset.Header {
		ix := i
		if col := normalise(h); col == "cardnumber" {
			index["cardnumber"] = ix + 1
			break
		}
	}

	for i, h := range recordset.Header {
		ix := i
		if col := normalise(h); col == "from" {
			index["startdate"] = ix + 1
			break
		}
	}

	for i, h := range recordset.Header {
		ix := i
		if col := normalise(h); col == "to" {
			index["enddate"] = ix + 1
			break
		}
	}

	for i, h := range recordset.Header {
		ix := i
		col := normalise(h)

		if col != "name" && col != "cardnumber" && col != "from" && col != "to" && col != "pin" {
			columns = append(columns, strings.ReplaceAll(h, " ", ""))
			index[col] = ix + 1
		}
	}

	for _, col := range columns {
		if index[normalise(col)] < 1 {
			return 0, fmt.Errorf("missing column %v", col)

		}
	}

	values := []string{}
	conflicts := []string{}
	for _, col := range columns {
		values = append(values, "?")

		if normalise(col) != "CardNumber" {
			conflicts = append(conflicts, fmt.Sprintf("%v=excluded.%v", col, col))
		}
	}

	sql := fmt.Sprintf("INSERT INTO %v (%v) VALUES (%v) ON CONFLICT(CardNumber) DO UPDATE SET %v;",
		table,
		strings.Join(columns, ","),
		strings.Join(values, ","),
		strings.Join(conflicts, ","))

	// ... execute
	count := 0

	if prepared, err := dbc.Prepare(sql); err != nil {
		return 0, err
	} else {
		for _, row := range recordset.Records {
			record := make([]any, len(columns))
			for i, col := range columns {
				ix := index[normalise(col)] - 1
				record[i] = row[ix]
			}

			if result, err := tx.Stmt(prepared).Exec(record...); err != nil {
				return 0, err
			} else if id, err := result.LastInsertId(); err != nil {
				return 0, err
			} else {
				count++
				debugf("put-acl: stored card %v@%v", row[index["cardnumber"]], id)
			}
		}
	}

	return count, nil
}
