package db

import (
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	lib "github.com/uhppoted/uhppoted-lib/acl"
)

type DB interface {
	GetACL(table string, withPIN bool) (*lib.Table, error)
	PutACL(table string, acl lib.Table, withPIN bool) (int, error)
	GetEvents(table string, controller uint32) ([]uint32, error)
	PutEvents(table string, events []core.Event) (int, error)
	AuditTrail(table string, trail []AuditRecord) (int, error)
	Log(table string, rs []LogRecord) (int, error)
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
