package cron

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

// A cron schedule that specifies a duty cycle (to the second granularity).
// Schedules are computed initially and stored as bit sets.
type Schedule struct {
	Second, Minute, Hour, Dom, Month, Dow uint64
}

// A range of acceptable values.
type bounds struct {
	min, max uint
	names    map[string]uint
}

// The bounds for each field.
var (
	seconds = bounds{0, 59, nil}
	minutes = bounds{0, 59, nil}
	hours   = bounds{0, 23, nil}
	dom     = bounds{1, 31, nil}
	months  = bounds{1, 12, map[string]uint{
		"jan": 1,
		"feb": 2,
		"mar": 3,
		"apr": 4,
		"may": 5,
		"jun": 6,
		"jul": 7,
		"aug": 8,
		"sep": 9,
		"oct": 10,
		"nov": 11,
		"dec": 12,
	}}
	dow = bounds{0, 7, map[string]uint{
		"sun": 0,
		"mon": 1,
		"tue": 2,
		"wed": 3,
		"thu": 4,
		"fri": 5,
		"sat": 6,
	}}
)

const (
	// Set the top bit if a star was included in the expression.
	STAR_BIT = 1 << 63
)

// Returns a new crontab entry representing the given spec.
// Panics with a descriptive error if the spec is not valid.
func Parse(spec string) *Schedule {
	if spec[0] == '@' {
		return parseDescriptor(spec)
	}

	// Split on whitespace.  We require 4 or 5 fields.
	// (minute) (hour) (day of month) (month) (day of week, optional)
	fields := strings.Fields(spec)
	fmt.Println("fields: ", fields)
	if len(fields) != 5 && len(fields) != 6 {
		log.Panicf("Expected 4 or 5 fields, found %d: %s", len(fields), spec)
	}

	// If a fifth field is not provided (DayOfWeek), then it is equivalent to star.
	if len(fields) == 5 {
		fields = append(fields, "*")
	}

	schedule := &Entry{
		Second: getField(fields[0], seconds),
		Minute: getField(fields[1], minutes),
		Hour:   getField(fields[2], hours),
		Dom:    getField(fields[3], dom),
		Month:  getField(fields[4], months),
		Dow:    getField(fields[5], dow),
	}

	// If either bit 0 or 7 are set, set both.  (both accepted as Sunday)
	if 1&schedule.Dow|1<<7&schedule.Dow > 0 {
		schedule.Dow = schedule.Dow | 1 | 1<<7
	}

	return schedule
}

// Return an Int with the bits set representing all of the times that the field represents.
// A "field" is a comma-separated list of "ranges".
func getField(field string, r bounds) uint64 {
	// list = range {"," range}
	var bits uint64
	fmt.Println("field: ", field)
	ranges := strings.FieldsFunc(field, func(r rune) bool { return r == ',' })
	fmt.Println("ranges: ", ranges)
	for _, expr := range ranges {
		bits |= getRange(expr, r)
	}
	return bits
}

func getRange(expr string, r bounds) uint64 {
	// number | number "-" number [ "/" number ]
	var (
		start, end, step uint
		rangeAndStep     = strings.Split(expr, "/")
		lowAndHigh       = strings.Split(rangeAndStep, "-")
		singleDigit      = len(lowAndHigh) == 1
	)

	if lowAndHigh[0] == "*" {
		start = r.min
		end = r.max
	} else {
		start = mustParseInt(lowAndHigh[0])
		switch len(lowAndHigh) {
		case 1:
			end = start
		case 2:
			fmt.Println("case 2: ", lowAndHigh[1])
			end = mustParseInt(lowAndHigh[1])
		default:
			log.Panicf("Too many commas: %s", expr)
		}
	}

	switch len(rangeAndStep) {
	case 1:
		step = 1
	case 2:
		step = mustParseInt(rangeAndStep[1])
	default:
		log.Panicf("Too many slashes: %s", expr)
	}
	fmt.Println("max ", start, end, step, r.min, r.max)

	if start < r.min {
		log.Panicf("Beginning of range (%d) below minimum (%d): %s", start, r.min, expr)
	}
	if end > r.max {
		log.Panicf("End of range (%d) above maximum (%d): %s", end, r.max, expr)
	}
	if start > end {
		log.Panicf("Beginning of range (%d) beyond end of range (%d): %s", start, end, expr)
	}

	return getBits(start, end, step)
}

func mustParseInt(expr string) uint {
	num, err := strconv.Atoi(expr)
	if err != nil {
		log.Panicf("Failed to parse int from %s: %s", expr, err)
	}
	if num < 0 {
		log.Panicf("Negative number (%d) not allowed: %s", num, expr)
	}

	return uint(num)
}

func getBits(min, max, step uint) uint64 {
	var bits uint64

	// If step is 1, use shifts.
	if step == 1 {
		fmt.Printf("max: %v, min: %v\n", max, min)
		fmt.Printf("%64b [^maxB]\n", uint64(^(math.MaxUint64 << (max + 1))))
		fmt.Printf("%64b [minB]\n", uint64(math.MaxUint64<<min))
		fmt.Printf("%08b [BBBBBB]\n", uint64(^(math.MaxUint64<<(max+1))&(math.MaxUint64<<min)))
		fmt.Printf("%v [uint64]\n\n", uint64(^(math.MaxUint64<<(max+1))&(math.MaxUint64<<min)))
		return ^(math.MaxUint64 << (max + 1)) & (math.MaxUint64 << min)
	}

	// Else, use a simple loop.
	for i := min; i <= max; i += step {
		fmt.Println("iiiii: ", i)
		fmt.Printf("%08b [bitsBB]\n", uint64(bits))
		bits |= 1 << i
	}
	fmt.Println("bits: ", bits)
	fmt.Println()
	return bits
}

func all(r Range) uint64 {
	return getBits(r.min, r.max, 1)
}

func first(r Range) uint64 {
	return getBits(r.min, r.min, 1)
}

func parseDescriptor(spec string) *Entry {
	switch spec {
	case "@yearly", "@annually":
		return &Entry{
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    1 << dom.min,
			Month:  1 << months.min,
			Dow:    all(dow),
		}

	case "@monthly":
		return &Entry{
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    1 << dom.min,
			Month:  all(months),
			Dow:    all(dow),
		}

	case "@weekly":
		return &Entry{
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    all(dom),
			Month:  all(months),
			Dow:    1 << dow.min,
		}

	case "@daily", "@midnight":
		return &Entry{
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    all(dom),
			Month:  all(months),
			Dow:    all(dow),
		}

	case "@hourly":
		return &Entry{
			Minute: 1 << minutes.min,
			Hour:   all(hours),
			Dom:    all(dom),
			Month:  all(months),
			Dow:    all(dow),
		}
	}

	log.Panicf("Unrecognized descriptor: %s", spec)
	return nil
}
