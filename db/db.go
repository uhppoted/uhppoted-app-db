package db

import (
	"time"

	lib "github.com/uhppoted/uhppoted-lib/acl"
)

type DB interface {
	GetACL(table string, withPIN bool) (*lib.Table, error)
}

type AuditRecord struct {
	Timestamp  time.Time
	Operation  string
	Controller uint32
	CardNumber uint32
	Status     string
	Card       string
}

type LogRecord struct {
	Timestamp  time.Time
	Operation  string
	Controller uint32
	Detail     string
}
