package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	lib "github.com/uhppoted/uhppoted-lib/acl"
)

func PutACL(dsn string, table string, recordset lib.Table, withPIN bool) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if dbc, err := open(dsn, MaxLifetime, MaxIdle, MaxOpen); err != nil {
		return 0, err
	} else if dbc == nil {
		return 0, fmt.Errorf("invalid PostgreSQL DB (%v)", dbc)
	} else if tx, err := dbc.BeginTx(ctx, nil); err != nil {
		return 0, err
	} else if count, err := insert(dbc, tx, table, recordset, withPIN); err != nil {
		return 0, err
	} else if err := tx.Commit(); err != nil {
		return 0, err
	} else {
		return count, nil
	}
}

func insert(dbc *sql.DB, tx *sql.Tx, table string, recordset lib.Table, withPIN bool) (int, error) {
	columns := []string{"CardNumber", "StartDate", "EndDate"}
	index := map[string]int{}

	for i, h := range recordset.Header {
		ix := i
		if col := normalise(h); col == "cardnumber" {
			index["cardnumber"] = ix + 1
			break
		}
	}

	if withPIN {
		columns = []string{"CardNumber", "PIN", "StartDate", "EndDate"}
		for i, h := range recordset.Header {
			ix := i
			if col := normalise(h); col == "pin" {
				index["pin"] = ix + 1
				break
			}
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
	for i, _ := range columns {
		values = append(values, fmt.Sprintf("$%d", i+1))
	}

	replace := []string{}
	for _, col := range columns {
		if normalise(col) != "cardnumber" {
			replace = append(replace, fmt.Sprintf("%v=EXCLUDED.%v", col, col))
		}
	}

	upsert := fmt.Sprintf("INSERT INTO %v (%v) VALUES (%v) ON CONFLICT(CardNumber) DO UPDATE SET %v;",
		table,
		strings.Join(columns, ","),
		strings.Join(values, ","),
		strings.Join(replace, ","))

	// ... execute
	count := 0

	if prepared, err := dbc.Prepare(upsert); err != nil {
		return 0, err
	} else {
		for _, row := range recordset.Records {
			card := row[index["cardnumber"]-1]
			record := []any{}
			for _, col := range columns {
				ix := index[normalise(col)] - 1
				column := normalise(col)

				if column == "cardnumber" {
					record = append(record, card)
				} else if column == "pin" {
					if row[ix] == "" {
						record = append(record, 0)
					} else if pin, err := strconv.ParseUint(row[ix], 10, 16); err != nil {
						return 0, err
					} else {
						record = append(record, pin)
					}
				} else if column == "startdate" {
					record = append(record, row[ix])
				} else if column == "enddate" {
					record = append(record, row[ix])
				} else {
					if row[ix] == "N" {
						record = append(record, 0)
					} else if row[ix] == "Y" {
						record = append(record, 1)
					} else {
						record = append(record, row[ix])
					}
				}
			}

			if _, err := tx.Stmt(prepared).Exec(record...); err != nil {
				return 0, err
			} else {
				count++
				debugf("put-acl: stored card %v", card)
			}
		}
	}

	return count, nil
}
