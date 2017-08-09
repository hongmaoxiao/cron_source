// This library implements a cron spec parser and runner.  See the README for
// more details.
package cron

import (
	"fmt"
	"sort"
	"time"
)

// Cron keeps track of any number of entries, invoking the associated func as
// specified by the spec.  See http://en.wikipedia.org/wiki/Cron
// It may be started and stopped.
type Cron struct {
	Entries []*Entry
	stop    chan struct{}
	add     chan *Entry
}

// A cron entry consists of a schedule and the func to execute on that schedule.
type Entry struct {
	*Schedule
	Next time.Time
	Func func()
}

type byTime []*Entry

func (s byTime) Len() int {
	return len(s)
}

func (s byTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byTime) Less(i, j int) bool {
	// Two zero times should return false.
	// Otherwise, zero is "greater" than any other time.
	// (To sort it at the end of the list.)
	if s[i].Next.IsZero() {
		return false
	}
	if s[j].Next.IsZero() {
		return true
	}
	return s[i].Next.Before(s[j].Next)
}

func New() *Cron {
	return &Cron{
		Entries: nil,
		add:     make(chan *Entry, 1),
		stop:    make(chan struct{}, 1),
	}
}

func (c *Cron) Add(spec string, cmd func()) {
	fmt.Println("before append: ", c.Entries)
	c.Entries = append(c.Entries, &Entry{Parse(spec), time.Time{}, cmd})
	fmt.Println("after append: ", c.Entries[0])
	entry := &Entry{Parse(spec), time.Time{}, cmd}
	select {
	case c.add <- entry:
		// The run loop accepted the entry, nothing more to do.
		fmt.Println("listening add: ", entry)
		return
	default:
		// No one listening to that channel, so just add to the array.
		fmt.Println("default append: ", c.Entries)
		c.Entries = append(c.Entries, entry)
		entry.Next = entry.Schedule.Next(time.Now().Local()) // Just in case...
	}

}

func (c *Cron) Start() {
	go c.Run()
}

func (c *Cron) Run() {
	// Figure out the next activation times for each entry.
	now := time.Now().Local()
	fmt.Println("now: ", now)
	fmt.Println("first entries: ", c.Entries[0])
	for _, entry := range c.Entries {
		fmt.Println("first in next: ")
		entry.Next = entry.Schedule.Next(now)
	}

	for {
		// Determine the next entry to run.
		fmt.Println("entry len: ", len(c.Entries))
		fmt.Println("before sort: ", c.Entries)
		sort.Sort(byTime(c.Entries))
		fmt.Println("after sort: ", c.Entries)

		var effective time.Time
		if len(c.Entries) == 0 || c.Entries[0].Next.IsZero() {
			// If there are no entries yet, just sleep - it still handles new entries
			// and stop requests.
			fmt.Println("no entries")
			effective = now.AddDate(10, 0, 0)
		} else {
			fmt.Println("entries: ", c.Entries[0])
			effective = c.Entries[0].Next
			fmt.Println("next time: ", effective)
		}

		select {
		case now = <-time.After(effective.Sub(now)):
			// Run every entry whose next time was this effective time.
			fmt.Println("in effective: ", c.Entries[0])
			for _, e := range c.Entries {
				fmt.Println("e.Next: ", e.Next)
				fmt.Println("effective: ", effective)
				if e.Next != effective {
					fmt.Println("not equare")
					break
				}
				fmt.Println("e.func", e.Func)
				go e.Func()
				e.Next = e.Schedule.Next(effective)
			}
		case newEntry := <-c.add:
			fmt.Println("in case add: ", newEntry)
			c.Entries = append(c.Entries, newEntry)
			newEntry.Next = newEntry.Schedule.Next(now)

		case <-c.stop:
			fmt.Println("in case stop")
			return
		}
	}
}

func (c Cron) Stop() {
	fmt.Println("cron stop")
	c.stop <- struct{}{}
}
