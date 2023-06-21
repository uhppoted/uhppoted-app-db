package db

import (
	"fmt"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
)

type AuditRecord struct {
	Timestamp  time.Time
	Operation  string
	Controller uint32
	CardNumber uint32
	Status     string
	Card       string
}

func NewAuditRecord(operation string, controller uint32, card core.Card, status string, withPIN bool) AuditRecord {
	return AuditRecord{
		Timestamp:  time.Now(),
		Operation:  operation,
		Controller: controller,
		CardNumber: card.CardNumber,
		Status:     status,
		Card:       format(card, withPIN),
	}
}

func door(door uint8, access uint8) string {
	switch access {
	case 0:
		return "N"
	case 1:
		return "Y"
	default:
		return fmt.Sprintf("%v", access)
	}
}

func format(c core.Card, withPIN bool) string {
	if withPIN {
		return fmt.Sprintf("%-10v %-10s %-10s %-3v %-3v %-3v %-3v %-5v",
			c.CardNumber,
			c.From,
			c.To,
			door(1, c.Doors[1]),
			door(2, c.Doors[2]),
			door(3, c.Doors[3]),
			door(4, c.Doors[4]),
			c.PIN)
	} else {
		return fmt.Sprintf("%-10v %-10s %-10s %-3v %-3v %-3v %-3v",
			c.CardNumber,
			c.From,
			c.To,
			door(1, c.Doors[1]),
			door(2, c.Doors[2]),
			door(3, c.Doors[3]),
			door(4, c.Doors[4]))
	}
}
