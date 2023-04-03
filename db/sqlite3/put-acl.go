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
	columns := []string{"CardNumber", "StartDate", "EndDate"}
	index := map[string]int{}

	for i, h := range table.Header {
		ix := i
		if col := normalise(h); col == "cardnumber" {
			index["cardnumber"] = ix + 1
			break
		}
	}

	for i, h := range table.Header {
		ix := i
		if col := normalise(h); col == "from" {
			index["startdate"] = ix + 1
			break
		}
	}

	for i, h := range table.Header {
		ix := i
		if col := normalise(h); col == "to" {
			index["enddate"] = ix + 1
			break
		}
	}

	for i, h := range table.Header {
		ix := i
		col := normalise(h)

		if col != "name" && col != "cardnumber" && col != "from" && col != "to" && col != "pin" {
			columns = append(columns, strings.ReplaceAll(h, " ", ""))
			index[col] = ix + 1
		}
	}

	for _, col := range columns {
		if index[normalise(col)] < 1 {
			return fmt.Errorf("missing column %v", col)

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
			record := make([]any, len(columns))
			for i, col := range columns {
				ix := index[normalise(col)] - 1
				record[i] = row[ix]
			}

			if _, err := prepared.ExecContext(ctx, record...); err != nil {
				return err
			}
		}
	}

	return nil
}
