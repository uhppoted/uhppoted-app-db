package mssql

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/microsoft/go-mssqldb"

	lib "github.com/uhppoted/uhppoted-lib/acl"
)

func PutACL(dsn string, recordset lib.Table, withPIN bool) (int, error) {
	table := "ACL"

	if match := regexp.MustCompile(`^(.*?)(::.*)$`).FindStringSubmatch(dsn); len(match) > 2 {
		dsn = match[1]
		table = match[2][2:]
	}

	// ... execute
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if dbc, err := open(dsn, MaxLifetime, MaxIdle, MaxOpen); err != nil {
		return 0, err
	} else if dbc == nil {
		return 0, fmt.Errorf("invalid %v DB (%v)", "SQL Server", dbc)
	} else if tx, err := dbc.BeginTx(ctx, nil); err != nil {
		return 0, err
	} else if _, err := clear(dbc, tx, table); err != nil {
		return 0, err
	} else if count, err := insert(dbc, tx, table, recordset, withPIN); err != nil {
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
	conflicts := []string{}
	for _, col := range columns {
		values = append(values, "?")

		if normalise(col) != "cardnumber" {
			conflicts = append(conflicts, fmt.Sprintf("%v=excluded.%v", col, col))
		}
	}

	// ... create all rows with card numbers, ignoring errors
	insert := fmt.Sprintf("INSERT INTO %v (CardNumber) VALUES (?);", table)

	if prepared, err := dbc.Prepare(insert); err != nil {
		return 0, err
	} else {
		for _, row := range recordset.Records {
			ix := index["cardnumber"] - 1
			card := row[ix]
			record := []any{card}

			if _, err := tx.Stmt(prepared).Exec(record...); err != nil {
				warnf("put-acl: error creating record for card number %v (%v)", card, err)
			} else {
				debugf("put-acl: created card record %v", card)
			}
		}
	}

	// ... update card number records with card information
	details := []string{}
	for _, col := range columns {
		if normalise(col) != "cardnumber" {
			details = append(details, fmt.Sprintf("%v=?", col))
		}
	}

	update := fmt.Sprintf("UPDATE %v SET %v WHERE CardNumber=?;", table, strings.Join(details, ","))

	// ... execute
	count := 0

	if prepared, err := dbc.Prepare(update); err != nil {
		return 0, err
	} else {
		for _, row := range recordset.Records {
			card := row[index["cardnumber"]-1]
			record := []any{}
			for _, col := range columns {
				ix := index[normalise(col)] - 1
				column := normalise(col)

				if column == "cardnumber" {
					continue
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

			record = append(record, card)

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
