package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/uhppoted/uhppoted-app-db/db"
)

func AuditTrail(dsn string, table string, recordset []db.AuditRecord) (int, error) {
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
	} else if N, err := appendToAuditTrail(dbc, tx, table, recordset); err != nil {
		return 0, err
	} else if err := tx.Commit(); err != nil {
		return 0, err
	} else {
		return N, nil
	}
}

func appendToAuditTrail(dbc *sql.DB, tx *sql.Tx, table string, recordset []db.AuditRecord) (int, error) {
	cardNumber := false
	card := false

	// ... get columns
	sql := fmt.Sprintf(`SELECT * FROM %v WHERE 1=2;`, table)
	if rs, err := tx.Query(sql); err != nil {
		return 0, err
	} else {
		defer rs.Close()

		if columns, err := rs.Columns(); err != nil {
			return 0, err
		} else {
			for _, col := range columns {
				if normalise(col) == "card" {
					card = true
				}

				if normalise(col) == "cardnumber" {
					cardNumber = true
				}
			}
		}
	}

	// ... append records
	count := 0
	insert := func() string {
		if card && cardNumber {
			return fmt.Sprintf("INSERT INTO %v (Operation, Controller, CardNumber, Status, Card) VALUES (?,?,?,?,?);", table)
		} else if card {
			return fmt.Sprintf("INSERT INTO %v (Operation, Controller, Status, Card) VALUES (?,?,?,?);", table)
		} else {
			return fmt.Sprintf("INSERT INTO %v (Operation, Controller, CardNumber, Status) VALUES (?,?,?,?);", table)
		}
	}

	g := func(r db.AuditRecord) []any {
		if card && cardNumber {
			return []any{r.Operation, r.Controller, r.CardNumber, r.Status, r.Card}
		} else if card {
			return []any{r.Operation, r.Controller, r.Status, r.Card}
		} else {
			return []any{r.Operation, r.Controller, r.CardNumber, r.Status}
		}
	}

	if prepared, err := dbc.Prepare(insert()); err != nil {
		return 0, err
	} else {
		for _, record := range recordset {
			row := g(record)

			if result, err := tx.Stmt(prepared).Exec(row...); err != nil {
				return 0, err
			} else if id, err := result.LastInsertId(); err != nil {
				return 0, err
			} else {
				count++
				debugf("audit: stored audit record for card %v@%v", record.CardNumber, id)
			}
		}
	}

	return count, nil
}
