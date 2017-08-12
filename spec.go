package cron

import (
	"fmt"
	"time"
)

// SpecSchedule specifies a duty cycle (to the second granularity), based on a
// traditional crontab specification. It is computed initially and stored as bit sets.
type SpecSchedule struct {
	Second, Minute, Hour, Dom, Month, Dow uint64
}

// bounds provides a range of acceptable values (plus a map of name to value).
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
	dow = bounds{0, 6, map[string]uint{
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
	starBit = 1 << 63
)

// Next returns the next time this schedule is activated, greater than the given
// time.  If no time can be found to satisfy the schedule, return the zero time.
func (s *SpecSchedule) Next(t time.Time) time.Time {
	// General approach:
	// For Month, Day, Hour, Minute, Second:
	// Check if the time value matches.  If yes, continue to the next field.
	// If the field doesn't match the schedule, then increment the field until it matches.
	// While incrementing the field, a wrap-around brings it back to the beginning
	// of the field list (since it is necessary to re-verify previous field
	// values)

	// Start at the earliest possible time (the upcoming second).
	t = t.Add(1*time.Second - time.Duration(t.Nanosecond())*time.Nanosecond)
	fmt.Println("added time: ", t)

	// This flag indicates whether a field has been incremented.
	added := false

	// If no time is found within five years, return zero.
	yearLimit := t.Year() + 5

WRAP:
	if t.Year() > yearLimit {
		return time.Time{}
	}

	// Find the first applicable month.
	// If it's this month, then do nothing.
	// fmt.Println("schedule: ", s)
	// fmt.Println("month: ", t.Month())
	// fmt.Println("t month: ", uint(t.Month()))
	// fmt.Printf("%64b: [64b]\n", 1<<uint(t.Month()))
	// fmt.Printf("%64b: [64b]\n", s.Month)
	// fmt.Printf("and Month: %v \n\n", 1<<uint(t.Month())&s.Month)
	for 1<<uint(t.Month())&s.Month == 0 {
		// If we have to add a month, reset the other parts to 0.
		// fmt.Println("in month")
		if !added {
			added = true
			// Otherwise, set the date at the beginning (since the current time is irrelevant).
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		}
		t = t.AddDate(0, 1, 0)
		// fmt.Println("added month: ", t)

		// Wrapped around.
		if t.Month() == time.January {
			goto WRAP
		}
	}

	// Now get a day in that month.
	for !dayMatches(s, t) {
		// fmt.Println("in day")
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		}
		t = t.AddDate(0, 0, 1)
		// fmt.Println("added day: ", t)

		if t.Day() == 1 {
			goto WRAP
		}
	}

	// fmt.Println("t hour", uint(t.Hour()))
	// fmt.Printf("%64b: [64b]\n", 1<<uint(t.Hour()))
	// fmt.Printf("%64b: [64b]\n", s.Hour)
	// fmt.Printf("%64b: [64b] \n", 1<<uint(t.Hour())&s.Hour)
	// fmt.Printf("and Hour: %v \n\n", 1<<uint(t.Hour())&s.Hour)
	for 1<<uint(t.Hour())&s.Hour == 0 {
		fmt.Println("in hour")
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
		}
		t = t.Add(1 * time.Hour)
		fmt.Println("added hour", t)

		if t.Hour() == 0 {
			goto WRAP
		}
	}

	// fmt.Println("t minute", uint(t.Minute()))
	// fmt.Printf("%64b: [64b]\n", 1<<uint(t.Minute()))
	// fmt.Printf("%64b: [64b]\n", s.Minute)
	// fmt.Printf("%64b: [64b] \n", 1<<uint(t.Minute())&s.Minute)
	// fmt.Printf("and Minute: %v \n\n", 1<<uint(t.Minute())&s.Minute)
	for 1<<uint(t.Minute())&s.Minute == 0 {
		fmt.Println("in minute")
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
		}
		t = t.Add(1 * time.Minute)
		// fmt.Println("added minute", t)

		if t.Minute() == 0 {
			goto WRAP
		}
	}

	// fmt.Println("t second", uint(t.Second()))
	// fmt.Printf("%64b: [64b]\n", 1<<uint(t.Second()))
	// fmt.Printf("%64b: [64b]\n", s.Second)
	// fmt.Printf("%64b: [64b] \n", 1<<uint(t.Second())&s.Second)
	// fmt.Printf("and Second: %v \n\n", 1<<uint(t.Second())&s.Second)
	for 1<<uint(t.Second())&s.Second == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
		}
		t = t.Add(1 * time.Second)
		// fmt.Println("added second", t)

		if t.Second() == 0 {
			goto WRAP
		}
	}

	fmt.Println("res: ", t)
	return t
}

// dayMatches returns true if the schedule's day-of-week and day-of-month
// restrictions are satisfied by the given time.
func dayMatches(s *SpecSchedule, t time.Time) bool {
	var (
		domMatch bool = 1<<uint(t.Day())&s.Dom > 0
		dowMatch bool = 1<<uint(t.Weekday())&s.Dow > 0
	)

	// fmt.Println("t dom", uint(t.Day()))
	// fmt.Printf("%64b: [64b]\n", 1<<uint(t.Day()))
	// fmt.Printf("%64b: [64b]\n", s.Dom)
	// fmt.Printf("%64b: [64b] \n", 1<<uint(t.Day())&s.Dom)
	// fmt.Printf("and dom: %v \n\n", 1<<uint(t.Day())&s.Dom)
	// fmt.Println("t Weekday", t.Weekday())
	// fmt.Println("t Weekday", uint(t.Weekday()))
	// fmt.Printf("%64b: [64b]\n", 1<<uint(t.Weekday()))
	// fmt.Printf("%64b: [64b]\n", s.Dow)
	// fmt.Printf("%64b: [64b] \n", 1<<uint(t.Weekday())&s.Dow)
	// fmt.Printf("and dow: %v \n\n", 1<<uint(t.Weekday())&s.Dow)
	// fmt.Printf("%64b: [64b]\n", uint(STAR_BIT))
	// fmt.Printf("%64b: [64b]\n", uint(s.Dom))
	// fmt.Printf("%64b: [64b]\n", uint(s.Dow))
	// fmt.Printf("s.Dom&STAR_BIT: %v \n", s.Dom&STAR_BIT)
	// fmt.Printf("s.Dow&STAR_BIT: %v \n\n", s.Dow&STAR_BIT)
	if s.Dom&starBit > 0 || s.Dow&starBit > 0 {
		// fmt.Println("iiiiiiiii")
		return domMatch && dowMatch
	}
	return domMatch || dowMatch
}
