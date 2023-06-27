package mysql

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/uhppoted/uhppoted-app-db/log"
)

const MaxLifetime = 5 * time.Minute
const MaxIdle = 2
const MaxOpen = 5
const LogTag = "mssql"

type record map[string]any

func open(dsn string, maxLifetime time.Duration, maxOpen int, maxIdle int) (*sql.DB, error) {
	dbc, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	dbc.SetConnMaxLifetime(maxLifetime)
	dbc.SetMaxOpenConns(maxOpen)
	dbc.SetMaxIdleConns(maxIdle)

	return dbc, nil
}

func row2record(rows *sql.Rows, columns []string, types []*sql.ColumnType) (record, error) {
	values := make([]any, len(types))
	pointers := make([]any, len(values))

	for i, v := range types {
		switch v.DatabaseTypeName() {
		case "VARCHAR":
			values[i] = ""

		case "DATE":
			values[i] = time.Time{}

		case "INT":
			values[i] = uint32(0)

		case "TINYINT":
			values[i] = uint8(0)

		default:
			return nil, fmt.Errorf("unsupported column type '%v'", v.DatabaseTypeName())
		}
	}

	for i := range values {
		pointers[i] = &values[i]
	}

	if err := rows.Scan(pointers...); err != nil {
		return nil, err
	} else {
		record := record{}

		for i := 0; i < len(columns); i++ {
			record[columns[i]] = values[i]
		}

		return record, nil
	}
}

func normalise(v string) string {
	return strings.ToLower(strings.ReplaceAll(v, " ", ""))
}

func clean(v string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(v), " ")
}

func debugf(format string, args ...interface{}) {
	f := fmt.Sprintf("%-10v %v", LogTag, format)

	log.Debugf(f, args...)
}

//lint:ignore U1000 utility function
func infof(format string, args ...interface{}) {
	f := fmt.Sprintf("%-10v %v", LogTag, format)

	log.Infof(f, args...)
}

func warnf(format string, args ...interface{}) {
	f := fmt.Sprintf("%-10v %v", LogTag, format)

	log.Warnf(f, args...)
}

//lint:ignore U1000 utility function
func errorf(format string, args ...any) {
	f := fmt.Sprintf("%-10v %v", LogTag, format)

	log.Errorf(f, args...)
}
