package mssql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
)

func GetEvents(dsn string, table string, controller uint32) ([]uint32, error) {
	query := fmt.Sprintf(`SELECT EventIndex FROM %v WHERE Controller=?;`, table)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if dbc, err := open(dsn, MaxLifetime, MaxIdle, MaxOpen); err != nil {
		return nil, err
	} else if dbc == nil {
		return nil, fmt.Errorf("invalid SQL Server DB (%v)", dbc)
	} else if prepared, err := dbc.Prepare(query); err != nil {
		return nil, err
	} else if rs, err := prepared.QueryContext(ctx, controller); err != nil {
		return nil, err
	} else if rs == nil {
		return nil, fmt.Errorf("invalid resultset (%v)", rs)
	} else {
		defer rs.Close()

		events := []uint32{}

		for rs.Next() {
			var event uint32
			if err := rs.Scan(&event); err != nil {
				return nil, err
			} else {
				events = append(events, event)
			}
		}

		return events, nil
	}
}

func PutEvents(dsn string, table string, events []core.Event) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if dbc, err := open(dsn, MaxLifetime, MaxIdle, MaxOpen); err != nil {
		return 0, err
	} else if dbc == nil {
		return 0, fmt.Errorf("invalid SQL Server DB (%v)", dbc)
	} else if tx, err := dbc.BeginTx(ctx, nil); err != nil {
		return 0, err
	} else if count, err := appendToEvents(dbc, tx, table, events); err != nil {
		return 0, err
	} else if err := tx.Commit(); err != nil {
		return 0, err
	} else {
		return count, nil
	}
}

func appendToEvents(dbc *sql.DB, tx *sql.Tx, table string, events []core.Event) (int, error) {
	count := 0
	insert := fmt.Sprintf("INSERT INTO %v (Controller,EventIndex) VALUES (?,?);", table)
	update := fmt.Sprintf("UPDATE %v SET Timestamp=?,Type=?,Granted=?,Door=?,Direction=?,CardNumber=?,Reason=? WHERE Controller=? AND EventIndex=?;", table)

	prepared := make([]*sql.Stmt, 2)

	if stmt, err := dbc.Prepare(insert); err != nil {
		return 0, err
	} else {
		prepared[0] = stmt
	}

	if stmt, err := dbc.Prepare(update); err != nil {
		return 0, err
	} else {
		prepared[1] = stmt
	}

	// ... create controller/event-index placeholder records (ignoring errors)
	for _, event := range events {
		tx.Stmt(prepared[0]).Exec(event.SerialNumber, event.Index)
	}

	// ... update placeholder records
	for _, event := range events {
		row := []any{
			fmt.Sprintf("%v", event.Timestamp),
			event.Type,
			event.Granted,
			event.Door,
			event.Direction,
			event.CardNumber,
			event.Reason,
			event.SerialNumber,
			event.Index,
		}

		if _, err := tx.Stmt(prepared[1]).Exec(row...); err != nil {
			return 0, err
		} else {
			count++
			debugf("stored event %v for %v", event.Index, event.SerialNumber)
		}
	}

	return count, nil
}
