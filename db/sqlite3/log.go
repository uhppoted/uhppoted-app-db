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

func Log(dsn string, table string, recordset []db.LogRecord) (int, error) {
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
		fmt.Sprintf("INSERT INTO %v (Operation, Detail) VALUES (?,?);", table),
		fmt.Sprintf("INSERT INTO %v (Operation, Controller, Detail) VALUES (?,?,?);", table),
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

			if result, err := tx.Stmt(prepared[0]).Exec(row...); err != nil {
				return 0, err
			} else if id, err := result.LastInsertId(); err != nil {
				return 0, err
			} else {
				count++
				debugf("log: stored operations record @%v", id)
			}
		} else {
			row := []any{record.Operation, record.Controller, record.Detail}

			if result, err := tx.Stmt(prepared[1]).Exec(row...); err != nil {
				return 0, err
			} else if id, err := result.LastInsertId(); err != nil {
				return 0, err
			} else {
				count++
				debugf("log: stored operations record for %v@%v", record.Controller, id)
			}
		}
	}

	return count, nil
}
