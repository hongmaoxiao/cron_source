package cron

import (
	"log"
	"strings"
)

type Entry struct {
	Minute, Hour, Dom, Month, Dow uint64
	Func                          func()
}

type Range struct{ min, max uint }

var (
	minutes = Range{0, 59}
	hours   = Range{0, 23}
	dom     = Range{1, 31}
	months  = Range{1, 12}
	dow     = Range{0, 7}
)

// Returns a new crontab entry representing the given spec.
// Panics with a descriptive error if the spec is not valid.
func NewEntry(spec string, cmd func()) *Entry {
	if spec[0] == '@' {
		entry := parseDescriptor(spec)
		entry.Func = cmd
		return entry
	}

	// Split on whitespace.  We require 4 or 5 fields.
	// (minute) (hour) (day of month) (month) (day of week, optional)
	fields := strings.Fields(spec)
	if len(fields) != 4 && len(fields) != 5 {
		log.Panicf("Expected 4 or 5 fields, found %d: %s", len(fields), spec)
	}

	entry := &Entry{
		Minute: getField(fields[0], minutes),
		Hour:   getField(fields[1], hours),
		Dom:    getField(fields[2], dom),
		Month:  getField(fields[3], months),
		Func:   cmd,
	}
	if len(fields) == 5 {
		entry.Dow = getField(fields[4], dow)

		// If either bit 0 or 7 are set, set both.  (both accepted as Sunday)
		if entry.Dow&1|entry.Dow&1<<7 > 0 {
			entry.Dow = entry.Dow | 1 | 1<<7
		}
	}

	return entry
}
