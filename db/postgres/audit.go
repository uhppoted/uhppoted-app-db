package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uhppoted/uhppoted-app-db/db"
)

func AuditTrail(dsn string, table string, recordset []db.AuditRecord) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if dbc, err := open(dsn, MaxLifetime, MaxIdle, MaxOpen); err != nil {
		return 0, err
	} else if dbc == nil {
		return 0, fmt.Errorf("invalid MySQL DB (%v)", dbc)
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
	if columns, err := getColumns(dbc, tx, table); err != nil {
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

	// ... append records
	count := 0
	insert := func() string {
		if card && cardNumber {
			return fmt.Sprintf("INSERT INTO %v (Operation, Controller, CardNumber, Status, Card) VALUES ($1,$2,$3,$4,$5);", table)
		} else if card {
			return fmt.Sprintf("INSERT INTO %v (Operation, Controller, Status, Card) VALUES ($1,$2,$3,$4);", table)
		} else {
			return fmt.Sprintf("INSERT INTO %v (Operation, Controller, CardNumber, Status) VALUES ($1,$2,$3,$4);", table)
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

			if _, err := tx.Stmt(prepared).Exec(row...); err != nil {
				return 0, err
			} else {
				count++
				debugf("audit: stored audit record for card %v", record.CardNumber)
			}
		}
	}

	return count, nil
}

func getColumns(dbc *sql.DB, tx *sql.Tx, table string) ([]string, error) {
	sql := fmt.Sprintf(`SELECT * FROM %v WHERE 1=2;`, table)

	if rs, err := tx.Query(sql); err != nil {
		return nil, err
	} else {
		defer rs.Close()

		if columns, err := rs.Columns(); err != nil {
			return nil, err
		} else {
			return columns, nil
		}
	}
}
