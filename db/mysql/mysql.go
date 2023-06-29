package mysql

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppoted-app-db/db"
	"github.com/uhppoted/uhppoted-app-db/log"
	lib "github.com/uhppoted/uhppoted-lib/acl"
)

const MaxLifetime = 5 * time.Minute
const MaxIdle = 2
const MaxOpen = 5
const LogTag = "mssql"

type record map[string]any

type dbi struct {
	dsn string
}

func NewDB(dsn string) db.DB {
	return dbi{
		dsn: dsn,
	}
}

func (d dbi) GetACL(table string, withPIN bool) (*lib.Table, error) {
	return GetACL(d.dsn, table, withPIN)
}

func (d dbi) PutACL(table string, acl lib.Table, withPIN bool) (int, error) {
	return PutACL(d.dsn, table, acl, withPIN)
}

func (d dbi) GetEvents(table string, controller uint32) ([]uint32, error) {
	return GetEvents(d.dsn, table, controller)
}

func (d dbi) PutEvents(table string, events []core.Event) (int, error) {
	return PutEvents(d.dsn, table, events)
}

func (d dbi) AuditTrail(table string, trail []db.AuditRecord) (int, error) {
	return AuditTrail(d.dsn, table, trail)
}

func (d dbi) Log(table string, rs []db.LogRecord) (int, error) {
	return Log(d.dsn, table, rs)
}

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

			// NTS: the MySQL driver always uses []uint8 rather than time.Time or sql.NullTime
		case "DATE":
			values[i] = make([]uint8, 16)

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

//lint:ignore U1000 utility function
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
