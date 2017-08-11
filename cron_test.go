package cron

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// Many tests schedule a job for every second, and then wait at most a second
// for it to run.  This amount is just slightly larger than 1 second to
// compensate for a few milliseconds of runtime.
const ONE_SECOND = 1*time.Second + 10*time.Millisecond

// // Start and stop cron with no entries.
// func TestNoEntries(t *testing.T) {
// cron := New()
// cron.Start()

// select {
// case <-time.After(ONE_SECOND):
// t.FailNow()
// case stop := <-stop(cron):
// fmt.Println("if stop: ", stop)
// }
// }

// // Start, stop, then add an entry. Verify entry doesn't run.
// func TestStopCausesJobsToNotRun(t *testing.T) {
// wg := &sync.WaitGroup{}
// wg.Add(1)

// cron := New()
// cron.Start()
// cron.Stop()
// cron.AddFunc("* * * * * ?", func() { wg.Done() })

// select {
// case <-time.After(ONE_SECOND):
// // No job ran!
// case <-wait(wg):
// t.FailNow()
// }
// }

// // Add a job, start cron, expect it runs.
// func TestAddBeforeRunning(t *testing.T) {
// wg := &sync.WaitGroup{}
// wg.Add(1)

// cron := New()
// cron.AddFunc("* * * * * ?", func() { wg.Done() })
// cron.Start()
// defer cron.Stop()

// // Give cron 2 seconds to run our job (which is always activated).
// select {
// case <-time.After(ONE_SECOND):
// t.FailNow()
// case <-wait(wg):
// }
// }

// Start cron, add a job, expect it runs.
func TestAddWhileRunning(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	cron := New()
	cron.Start()
	defer cron.Stop()
	cron.AddFunc("* * * * * ?", func() { wg.Done() })

	select {
	case <-time.After(ONE_SECOND):
		t.FailNow()
	case <-wait(wg):
	}
}

// // Test that the entries are correctly sorted.
// // Add a bunch of long-in-the-future entries, and an immediate entry, and ensure
// // that the immediate entry runs immediately.
// // Also: Test that multiple jobs run in the same instant.
// func TestMultipleEntries(t *testing.T) {
// wg := &sync.WaitGroup{}
// wg.Add(2)

// cron := New()
// cron.AddFunc("0 0 0 1 1 ?", func() {})
// cron.AddFunc("* * * * * ?", func() {
// wg.Done()
// })
// cron.AddFunc("0 0 0 31 12 ?", func() {})
// cron.AddFunc("* * * * * ?", func() {
// wg.Done()
// })

// cron.Start()
// defer cron.Stop()

// select {
// case <-time.After(ONE_SECOND):
// t.FailNow()
// case <-wait(wg):
// }
// }

// // Test running the same job twice.
// func TestRunningJobTwice(t *testing.T) {
// wg := &sync.WaitGroup{}
// wg.Add(2)

// cron := New()
// cron.AddFunc("0 0 0 1 1 ?", func() {})
// cron.AddFunc("* * * * * ?", func() {
// wg.Done()
// })
// cron.AddFunc("0 0 0 31 12 ?", func() {})

// cron.Start()
// defer cron.Stop()

// select {
// case <-time.After(2 * ONE_SECOND):
// t.FailNow()
// case <-wait(wg):
// }
// }

// // Test that the cron is run in the local time zone (as opposed to UTC).
// func TestLocalTimezone(t *testing.T) {
// wg := &sync.WaitGroup{}
// wg.Add(1)

// now := time.Now().Local()
// spec := fmt.Sprintf("%d %d %d %d %d ?", now.Second()+1, now.Minute(), now.Hour(), now.Day(), now.Month())

// cron := New()
// cron.AddFunc(spec, func() { wg.Done() })
// cron.Start()
// defer cron.Stop()

// select {
// case <-time.After(ONE_SECOND):
// t.FailNow()
// case <-wait(wg):
// }
// }

// type testJob struct {
// wg   *sync.WaitGroup
// name string
// }

// func (t testJob) Run() {
// t.wg.Done()
// }

// // Simple test using Runnables.
// func TestJob(t *testing.T) {
// wg := &sync.WaitGroup{}
// wg.Add(1)

// cron := New()
// cron.AddJob("0 0 0 30 Feb ?", testJob{wg, "job0"})
// cron.AddJob("0 0 0 1 1 ?", testJob{wg, "job1"})
// cron.AddJob("* * * * * ?", testJob{wg, "job2"})
// cron.AddJob("1 0 0 1 1 ?", testJob{wg, "job3"})

// cron.Start()
// defer cron.Stop()

// select {
// case <-time.After(ONE_SECOND):
// t.FailNow()
// case <-wait(wg):
// }

// // Ensure the entries are in the right order.
// answers := []string{"job2", "job1", "job3", "job0"}
// for i, answer := range answers {
// actual := cron.Entries()[i].Job.(testJob).name
// fmt.Println("actual in turn: ", actual)
// if actual != answer {
// t.Errorf("Jobs not in the right order.  (expected) %s != %s (actual)", answer, actual)
// }
// }
// }

func wait(wg *sync.WaitGroup) chan bool {
	ch := make(chan bool)
	go func() {
		fmt.Println("wg wait")
		wg.Wait()
		ch <- true
	}()
	return ch
}

func stop(cron *Cron) chan bool {
	ch := make(chan bool)
	go func() {
		cron.Stop()
		ch <- true
	}()
	return ch
}
