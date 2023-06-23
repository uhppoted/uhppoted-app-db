package db

import (
	"time"
)

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
