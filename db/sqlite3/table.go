package sqlite3

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	lib "github.com/uhppoted/uhppoted-lib/acl"
)

func makeTable(columns []string, recordset []record, withPIN bool) (*lib.Table, error) {
	if len(recordset) == 0 {
		return nil, fmt.Errorf("empty ACL table")
	}

	// .. build index
	index := map[string]string{}
	for _, v := range columns {
		k := normalise(v)
		if _, ok := index[k]; ok {
			return nil, fmt.Errorf("duplicate column name '%s'", v)
		}

		index[k] = v
	}

	// ... build header
	header := []string{}

	if _, ok := index["cardnumber"]; !ok {
		return nil, fmt.Errorf("missing 'card number' column")
	} else {
		header = append(header, "Card Number")
	}

	if withPIN {
		if _, ok := index["pin"]; !ok {
			return nil, fmt.Errorf("missing 'PIN' column")
		} else {
			header = append(header, "PIN")
		}
	}

	if _, ok := index["startdate"]; !ok {
		return nil, fmt.Errorf("missing 'from' column")
	} else {
		header = append(header, "From")
	}

	if _, ok := index["enddate"]; !ok {
		return nil, fmt.Errorf("missing 'to' column")
	} else {
		header = append(header, "To")
	}

	for _, v := range columns {
		k := normalise(v)
		if k != "cardnumber" && k != "startdate" && k != "enddate" && k != "name" && k != "pin" {
			header = append(header, clean(v))
		}
	}

	// ... records
	rows := [][]string{}

	for _, record := range recordset {
		row := []string{}

		if cardnumber, ok := record[index["cardnumber"]].(int64); !ok {
			continue
		} else if cardnumber < 0 || cardnumber > math.MaxUint32 {
			warnf("sqlite3", "invalid card number (%v)", cardnumber)
			continue
		} else {
			row = append(row, fmt.Sprintf("%v", uint32(cardnumber)))
		}

		if withPIN {
			if pin, ok := record[index["pin"]].(int64); !ok {
				continue
			} else if pin < 0 || pin > math.MaxUint16 {
				warnf("sqlite3", "invalid PIN (%v)", pin)
				continue
			} else {
				row = append(row, fmt.Sprintf("%v", uint32(pin)))
			}
		}

		if from, ok := record[index["startdate"]].(string); !ok {
			continue
		} else if t, err := time.ParseInLocation("2006-01-02", from, time.Local); err != nil {
			warnf("sqlite3", "invalid start date (%v)", from)
			continue
		} else {
			row = append(row, t.Format("2006-01-02"))
		}

		if to, ok := record[index["enddate"]].(string); !ok {
			continue
		} else if t, err := time.ParseInLocation("2006-01-02", to, time.Local); err != nil {
			warnf("sqlite3", "invalid end date (%v)", to)
			continue
		} else {
			row = append(row, t.Format("2006-01-02"))
		}

		var doors []string

		if withPIN {
			doors = header[4:]
		} else {
			doors = header[3:]
		}

		for _, h := range doors {
			if k, ok := index[normalise(h)]; ok {
				if permission, ok := record[k].(int64); !ok {
					row = append(row, "")
				} else if permission == 0 {
					row = append(row, "N")
				} else if permission == 1 {
					row = append(row, "Y")
				} else if permission < 255 {
					row = append(row, fmt.Sprintf("%v", permission))
				} else {
					row = append(row, "")
				}
			} else {
				row = append(row, "")
			}
		}

		rows = append(rows, row)
	}

	return &lib.Table{
		Header:  header,
		Records: rows,
	}, nil
}

func normalise(v string) string {
	return strings.ToLower(strings.ReplaceAll(v, " ", ""))
}

func clean(v string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(v), " ")
}
