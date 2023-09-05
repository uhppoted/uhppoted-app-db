package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uhppoted/uhppoted-app-db/db"
)

func Log(dsn string, table string, recordset []db.LogRecord) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if dbc, err := open(dsn, MaxLifetime, MaxIdle, MaxOpen); err != nil {
		return 0, err
	} else if dbc == nil {
		return 0, fmt.Errorf("invalid PostgreSQL DB (%v)", dbc)
	} else if tx, err := dbc.BeginTx(ctx, nil); err != nil {
		return 0, err
	} else if N, err := appendToLog(dbc, tx, table, recordset); err != nil {
		return 0, err
	} else if err := tx.Commit(); err != nil {
		return 0, err
	} else {
		return N, nil
	}
}

func appendToLog(dbc *sql.DB, tx *sql.Tx, table string, recordset []db.LogRecord) (int, error) {
	count := 0
	insert := []string{
		fmt.Sprintf("INSERT INTO %v (Operation, Detail) VALUES ($1,$2);", table),
		fmt.Sprintf("INSERT INTO %v (Operation, Controller, Detail) VALUES ($1,$2,$3);", table),
	}

	prepared := []*sql.Stmt{}

	for _, sql := range insert {
		if stmt, err := dbc.Prepare(sql); err != nil {
			return 0, err
		} else {
			prepared = append(prepared, stmt)
		}
	}

	for _, record := range recordset {
		if record.Controller == 0 {
			row := []any{record.Operation, record.Detail}

			if _, err := tx.Stmt(prepared[0]).Exec(row...); err != nil {
				return 0, err
			} else {
				count++
				debugf("log: stored operations record")
			}
		} else {
			row := []any{record.Operation, record.Controller, record.Detail}

			if _, err := tx.Stmt(prepared[1]).Exec(row...); err != nil {
				return 0, err
			} else {
				count++
				debugf("log: stored operations record for %v", record.Controller)
			}
		}
	}

	return count, nil
}
