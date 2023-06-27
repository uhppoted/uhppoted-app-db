package commands

import (
	"flag"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	core "github.com/uhppoted/uhppote-core/types"
	"github.com/uhppoted/uhppote-core/uhppote"
	"github.com/uhppoted/uhppoted-app-db/db"
	"github.com/uhppoted/uhppoted-app-db/log"
	"github.com/uhppoted/uhppoted-lib/config"
)

type interval struct {
	from uint32
	to   uint32
}

func (i interval) contains(v uint32) bool {
	return i.from <= v && i.to >= v
}

const BATCHSIZE = 5 // maximum controller events to fetch
const GAPS = 2      // fix at most 2 gaps in a controller event record

var GetEventsCmd = GetEvents{
	command: command{
		name:        "get-events",
		description: "Retrieves a batch of events from the set of configured controllers and stores the events to a database table",
		usage:       "--dsn <DSN> [--table:events <table>] [-table:log <table>] [--batch-size <N>]",

		dsn: "",
		tables: tables{
			Events: "events",
			Log:    "",
		},
		lockfile: "",
		config:   config.DefaultConfig,
		debug:    false,
	},
	batchSize: BATCHSIZE,
}

type GetEvents struct {
	command
	batchSize uint
}

func (cmd *GetEvents) Help() {
	fmt.Println()
	fmt.Printf("  Usage: %s [--debug] [--config <file>] get-events --dsn <DSN> [--table:events <table>] [-table:log <table>] [--batch-size <N>]\n", APP)
	fmt.Println()
	fmt.Println("  Retrieves a batch of events from the set of configured controllers and adds the events to the events table")
	fmt.Println()

	helpOptions(cmd.FlagSet())

	fmt.Println()
	fmt.Println("  Examples:")
	fmt.Println(`    uhppote-app-db --debug get-events --dsn "sqlite3://./db/ACL.db`)
	fmt.Println(`    uhppote-app-db --debug get-events --dsn "sqlite3://./db/ACL.db" --table:events  events -table:log OpsLog --batch-size 500"`)
	fmt.Println()
}

func (cmd *GetEvents) FlagSet() *flag.FlagSet {
	flagset := flag.NewFlagSet("get-events", flag.ExitOnError)

	flagset.StringVar(&cmd.dsn, "dsn", cmd.dsn, "DSN for database")
	flagset.StringVar(&cmd.tables.Events, "table:events", cmd.tables.Events, "Events table name. Defaults to 'events'")
	flagset.StringVar(&cmd.tables.Log, "table:log", cmd.tables.Log, "Operations log table name. Defaults to ''")
	flagset.UintVar(&cmd.batchSize, "batch-size", cmd.batchSize, "Maximum events (per controller) to retrieve per invocation. Defaults to 100.")
	flagset.StringVar(&cmd.lockfile, "lockfile", cmd.lockfile, "Filepath for lock file. Defaults to <tmp>/uhppoted-app-db.lock")

	return flagset
}

func (cmd *GetEvents) Execute(args ...any) error {
	options := args[0].(*Options)

	cmd.config = options.Config
	cmd.debug = options.Debug

	log.SetDebug(options.Debug)

	// ... check parameters
	if strings.TrimSpace(cmd.dsn) == "" {
		return fmt.Errorf("invalid database DSN")
	}

	if strings.TrimSpace(cmd.tables.Events) == "" {
		return fmt.Errorf("invalid events table")
	}

	// ... locked?
	if kraken, err := lock(cmd.lockfile); err != nil {
		return err
	} else {
		defer func() {
			infof("get-events", "removing lockfile")
			kraken.Release()
		}()
	}

	// ... get config
	conf := config.NewConfig()
	if err := conf.Load(cmd.config); err != nil {
		return fmt.Errorf("could not load configuration (%v)", err)
	}

	u, devices := getDevices(conf, false)

	// ... retrieve events from controllers
	events := []core.Event{}
	errors := []error{}

	for _, device := range devices {
		controller := device.DeviceID

		if list, err := cmd.getEvents(u, controller); err != nil {
			warnf("get-events", "%v  %v", controller, err)
			errors = append(errors, err)
		} else {
			events = append(events, list...)
		}
	}

	// ... store to DB
	if err := putEvents(cmd.dsn, cmd.tables.Events, events); err != nil {
		return err
	}

	// ... add operations log
	if cmd.tables.Log != "" {
		recordset := []db.LogRecord{
			db.LogRecord{
				Timestamp: time.Now(),
				Operation: "get-events",
				Detail:    fmt.Sprintf("records:%v  errors:%v", len(events), len(errors)),
			},
		}

		if err := stashToLog(cmd.dsn, cmd.tables.Log, recordset); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *GetEvents) getEvents(u uhppote.IUHPPOTE, controller uint32) ([]core.Event, error) {
	infof("get-events", "%v  retrieving events", controller)

	if first, last, current, err := getEventIndices(u, controller); err != nil {
		return nil, err
	} else {
		debugf("get-events", "%v  first:%-6v last:%-6v current:%-6v\n", controller, first, last, current)

		var intervals []interval
		if list, err := cmd.getMissing(GAPS, controller); err != nil {
			return nil, err
		} else {
			intervals = list
		}

		for _, interval := range intervals {
			if interval.contains(last) || interval.contains(first) || (interval.from >= first && interval.to <= last) {
				break
			}
		}

		count := uint(0)
		events := []core.Event{}

		f := func(index uint32) {
			if e, err := u.GetEvent(controller, index); err != nil {
				warnf("get-events", "%v %v", controller, err)
			} else if e == nil {
				warnf("get-events", "%v  missing event %v", controller, index)
				events = append(events, core.Event{
					SerialNumber: core.SerialNumber(controller),
					Index:        index,
				})
			} else {
				events = append(events, core.Event{
					Timestamp:    e.Timestamp,
					SerialNumber: core.SerialNumber(controller),
					Index:        e.Index,
					Type:         e.Type,
					Granted:      e.Granted,
					Door:         e.Door,
					Direction:    e.Direction,
					CardNumber:   e.CardNumber,
					Reason:       e.Reason,
				})
			}

			count++
		}

		for _, interval := range intervals {
			if interval.contains(last) {
				index := interval.from
				if index < first {
					index = first
				}

				for index <= last && count < cmd.batchSize {
					f(index)
					index++
				}
			}

			if interval.contains(first) {
				index := interval.to
				if index > last {
					index = last
				}

				for index >= first && count < cmd.batchSize {
					f(index)
					index--
				}
			}

			if interval.from >= first && interval.to <= last {
				for index := interval.from; index <= interval.to; index++ {
					f(index)
				}
			}

		}

		infof("get-events", "retrieved %v events", count)

		return events, nil
	}
}

func (cmd *GetEvents) getMissing(gaps int, controller uint32) ([]interval, error) {
	var events []uint32
	var intervals []interval

	if list, err := getEvents(cmd.dsn, cmd.tables.Events, controller); err != nil {
		return nil, err
	} else {
		events = list
	}

	sort.Slice(events, func(i, j int) bool { return events[i] < events[j] })

	first := uint32(0)
	last := uint32(0)

	if N := len(events); N > 0 {
		first = events[0]
		last = events[N-1]
	}

	intervals = append(intervals, interval{from: last + 1, to: math.MaxUint32})
	if first > 1 {
		intervals = append(intervals, interval{from: 1, to: first - 1})
	}

	slice := events[0:]
	for len(slice) > 0 && gaps != 0 {
		ix := sort.Search(len(slice), func(i int) bool {
			return slice[i] != slice[0]+uint32(i)
		})

		if ix != len(slice) {
			from := slice[ix-1] + 1
			to := slice[ix] - 1
			intervals = append(intervals, interval{from: from, to: to})
			gaps--
		}

		slice = slice[ix:]
	}

	return intervals, nil
}

func getEventIndices(u uhppote.IUHPPOTE, controller uint32) (uint32, uint32, uint32, error) {
	var first uint32 = 0
	var last uint32 = 0
	var current uint32 = 0

	if v, err := u.GetEvent(controller, 0); err != nil {
		return 0, 0, 0, err
	} else if v != nil {
		first = v.Index
	}

	if v, err := u.GetEvent(controller, 0xffffffff); err != nil {
		return 0, 0, 0, err
	} else if v != nil {
		last = v.Index
	}

	if v, err := u.GetEventIndex(controller); err != nil {
		return 0, 0, 0, err
	} else if v != nil {
		current = v.Index
	}

	return first, last, current, nil
}
