package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
)

func GetEvents(dsn string, table string, controller uint32) ([]uint32, error) {
	query := fmt.Sprintf(`SELECT EventIndex FROM %v WHERE Controller=$1;`, table)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if dbc, err := open(dsn, MaxLifetime, MaxIdle, MaxOpen); err != nil {
		return nil, err
	} else if dbc == nil {
		return nil, fmt.Errorf("invalid PostgreSQL DB (%v)", dbc)
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
		return 0, fmt.Errorf("invalid PostgreSQL DB (%v)", dbc)
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

	columns := []string{"Timestamp", "Controller", "EventIndex", "Type", "Granted", "Door", "Direction", "CardNumber", "Reason"}
	values := []string{"$1", "$2", "$3", "$4", "$5", "$6", "$7", "$8", "$9"}
	replace := []string{
		"Timestamp=EXCLUDED.Timestamp",
		"Type=EXCLUDED.Type",
		"Granted=EXCLUDED.Granted",
		"Door=EXCLUDED.Door",
		"Direction=EXCLUDED.Direction",
		"CardNumber=EXCLUDED.CardNumber",
		"Reason=EXCLUDED.Reason",
	}

	upsert := fmt.Sprintf("INSERT INTO %v (%v) VALUES (%v) ON CONFLICT ON CONSTRAINT ControllerEventIndex DO UPDATE SET %v;",
		table,
		strings.Join(columns, ","),
		strings.Join(values, ","),
		strings.Join(replace, ","))

	if prepared, err := dbc.Prepare(upsert); err != nil {
		return 0, err
	} else {
		granted := func(v bool) uint8 {
			if v {
				return 1
			} else {
				return 0
			}
		}

		for _, event := range events {
			row := []any{
				fmt.Sprintf("%v", event.Timestamp),
				event.SerialNumber,
				event.Index,
				event.Type,
				granted(event.Granted),
				event.Door,
				event.Direction,
				event.CardNumber,
				event.Reason,
			}

			if _, err := tx.Stmt(prepared).Exec(row...); err != nil {
				return 0, err
			} else {
				count++
				debugf("stored event %v for %v", event.Index, event.SerialNumber)
			}
		}
	}

	return count, nil
}
