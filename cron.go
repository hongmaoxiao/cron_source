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
	entries  []*Entry
	stop     chan struct{}
	add      chan *Entry
	snapshot chan []*Entry
	running  bool
}

// Simple interface for submitted cron jobs.
type Job interface {
	Run()
}

// A cron entry consists of a schedule and the func to execute on that schedule.
type Entry struct {
	*Schedule
	Next time.Time
	Prev time.Time
	Job  Job
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
		entries:  nil,
		add:      make(chan *Entry),
		stop:     make(chan struct{}),
		snapshot: make(chan []*Entry),
		running:  false,
	}
}

// Provide a default implementation for those that want to run a simple func.
type jobAdapter func()

func (r jobAdapter) Run() {
	r()
}

func (c *Cron) AddFunc(spec string, cmd func()) {
	c.AddJob(spec, jobAdapter(cmd))
}

func (c *Cron) AddJob(spec string, cmd Job) {
	fmt.Println("before append entry len: ", len(c.entries))
	entry := &Entry{
		Schedule: Parse(spec)
		Job: cmd,
	}
	if !c.running {
		fmt.Println("not running, append entries")
		c.entries = append(c.entries, entry)
		fmt.Println("after append entry len: ", len(c.entries))
		return
	}

	c.add <- entry
}

// Return a snapshot of the cron entries.
func (c *Cron) Entries() []*Entry {
	if c.running {
		fmt.Println("snapshot")
		c.snapshot <- nil
		fmt.Println("block......")
		x := <-c.snapshot
		fmt.Println("x: ", x)
		return x
	}
	return c.entrySnapshot()
}

func (c *Cron) Start() {
	fmt.Println("cron start")
	c.running = true
	go c.run()
}

// Run the scheduler.. this is private just due to the need to synchronize
// access to the 'running' state variable.
func (c *Cron) run() {
	// Figure out the next activation times for each entry.
	now := time.Now().Local()
	fmt.Println("now: ", now)
	fmt.Println("first entries: ", c.entries)
	for _, entry := range c.entries {
		fmt.Println("range for entry: ", entry)
		fmt.Println("first in next: ")
		entry.Next = entry.Schedule.Next(now)
	}

	for {
		// Determine the next entry to run.
		fmt.Println("entry len: ", len(c.entries))
		fmt.Println("before sort: ", c.entries)
		sort.Sort(byTime(c.entries))
		fmt.Println("after sort: ", c.entries)

		var effective time.Time
		if len(c.entries) == 0 || c.entries[0].Next.IsZero() {
			// If there are no entries yet, just sleep - it still handles new entries
			// and stop requests.
			fmt.Println("no entries")
			effective = now.AddDate(10, 0, 0)
		} else {
			fmt.Println("entries 0: ", c.entries[0])
			effective = c.entries[0].Next
			fmt.Println("next time: ", effective)
		}

		select {
		case now = <-time.After(effective.Sub(now)):
			// Run every entry whose next time was this effective time.
			fmt.Println("arrival now: ", now)
			fmt.Println("in effective: ", c.entries[0])
			for _, e := range c.entries {
				fmt.Println("e.Next: ", e.Next)
				fmt.Println("effective: ", effective)
				if e.Next != effective {
					fmt.Println("not equare")
					break
				}
				fmt.Println("e.func", e.Job)
				go e.Job.Run()
				e.Prev = e.Next
				e.Next = e.Schedule.Next(effective)
			}
		case newEntry := <-c.add:
			fmt.Println("in case add: ", newEntry)
			c.entries = append(c.entries, newEntry)
			newEntry.Next = newEntry.Schedule.Next(now)

		case sn := <-c.snapshot:
			fmt.Println("receive snapshot: ", sn)
			c.snapshot <- c.entrySnapshot()

		case <-c.stop:
			fmt.Println("in case stop")
			return
		}
	}
}

func (c *Cron) Stop() {
	fmt.Println("cron stop")
	c.stop <- struct{}{}
	c.running = false
}

func (c *Cron) entrySnapshot() []*Entry {
	entries := []*Entry{}
	for _, e := range c.entries {
		entries = append(entries, &Entry{
			Schedul: e.Schedule,
			Next: e.Next,
			Prev: e.Prev,
			Job: e.Job,
		})
	}
	return entries
}
