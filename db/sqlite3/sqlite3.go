package sqlite3

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	// "github.com/uhppoted/uhppoted-app-db/log"
)

const MaxLifetime = 5 * time.Minute
const MaxIdle = 2
const MaxOpen = 5
const LogTag = "sqlite3"

func GetACL(dsn string) error {
	if _, err := os.Stat(dsn); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("sqlite3 database %v does not exist", dsn)
	} else if err != nil {
		return err
	}

	if dbc, err := open(dsn, MaxLifetime, MaxIdle, MaxOpen); err != nil {
		return err
	} else if dbc == nil {
		return fmt.Errorf("invalid sqlite3 DB pool (%v)", dbc)
	} else if prepared, err := dbc.Prepare(sqlAclGet); err != nil {
		return err
	} else {
		fmt.Printf(">> %v\n", prepared)
	}

	// return &dbx{
	//     pool:     dbc,
	//     prepared: prepared,
	// }, nil

	return nil
}

func open(path string, maxLifetime time.Duration, maxOpen int, maxIdle int) (*sql.DB, error) {
	dbc, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	dbc.SetConnMaxLifetime(maxLifetime)
	dbc.SetMaxOpenConns(maxOpen)
	dbc.SetMaxIdleConns(maxIdle)

	return dbc, nil
}

// func debugf(format string, args ...interface{}) {
// 	f := fmt.Sprintf("%-10v %v", LogTag, format)
//
// 	log.Debugf(f, args...)
// }

// func infof(format string, args ...interface{}) {
// 	f := fmt.Sprintf("%-10v %v", LogTag, format)
//
// 	log.Infof(f, args...)
// }

// func warnf(format string, args ...interface{}) {
// 	f := fmt.Sprintf("%-10v %v", LogTag, format)
//
// 	log.Warnf(f, args...)
// }

// func errorf(format string, args ...any) {
// 	f := fmt.Sprintf("%-10v %v", LogTag, format)
//
// 	log.Errorf(f, args...)
// }
