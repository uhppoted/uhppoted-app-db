package postgres

import (
	"fmt"
	"math"
	"strconv"
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
			warnf("mssql", "invalid card number (%v)", cardnumber)
			continue
		} else {
			row = append(row, fmt.Sprintf("%v", uint32(cardnumber)))
		}

		if withPIN {
			if pin, ok := record[index["pin"]].(int64); !ok {
				continue
			} else if pin < 0 || pin > math.MaxUint16 {
				warnf("mssql", "invalid PIN (%v)", pin)
				println("oops/4")
				continue
			} else {
				row = append(row, fmt.Sprintf("%v", uint32(pin)))
			}
		}

		// NTS: bizarrely, the MySQL driver converts a time.Time value to []uint8
		if from, ok := record[index["startdate"]].(time.Time); ok {
			row = append(row, from.Format("2006-01-02"))
		} else if from, ok := record[index["startdate"]].([]uint8); ok {
			row = append(row, string(from))
		} else {
			continue
		}

		// NTS: bizarrely, the MySQL driver converts a time.Time value to []uint8
		if to, ok := record[index["enddate"]].(time.Time); ok {
			row = append(row, to.Format("2006-01-02"))
		} else if to, ok := record[index["enddate"]].([]uint8); ok {
			row = append(row, string(to))
		} else {
			continue
		}

		var doors []string

		if withPIN {
			doors = header[4:]
		} else {
			doors = header[3:]
		}

		for _, h := range doors {
			if k, ok := index[normalise(h)]; !ok {
				row = append(row, "")
			} else if s, ok := record[k].(string); ok {
				if s == "N" || s == "n" {
					row = append(row, "N")
				} else if s == "Y" || s == "y" {
					row = append(row, "Y")
				} else if v, err := strconv.ParseUint(s, 10, 8); err == nil {
					row = append(row, fmt.Sprintf("%v", v))
				} else {
					row = append(row, "")
				}

			} else if permission, ok := record[k].(int64); ok {
				if permission == 0 {
					row = append(row, "N")
				} else if permission == 1 {
					row = append(row, "Y")
				} else if permission < 255 {
					row = append(row, fmt.Sprintf("%v", permission))
				} else {
					row = append(row, "")
				}
			}
		}

		rows = append(rows, row)
	}

	return &lib.Table{
		Header:  header,
		Records: rows,
	}, nil
}
