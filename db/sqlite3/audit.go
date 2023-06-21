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
	count := 0
	sql := fmt.Sprintf("INSERT INTO %v (Operation, Controller, CardNumber, Status, Card) VALUES (?,?,?,?,?);", table)

	if prepared, err := dbc.Prepare(sql); err != nil {
		return 0, err
	} else {
		for _, record := range recordset {
			row := []any{
				record.Operation,
				record.Controller,
				record.CardNumber,
				record.Status,
				record.Card,
			}

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
