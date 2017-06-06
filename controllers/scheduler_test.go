package controllers

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func schedulerTest(f func(t *testing.T, s *Scheduler)) (string, func(*testing.T)) {
	return testFunctionName(f), func(t *testing.T) {
		t.Parallel()
		s := NewScheduler()
		f(t, s)
		s.CancelAll()
	}
}

func TestScheduler(t *testing.T) {
	t.Run("Scheduler", func(t *testing.T) {
		t.Parallel()
		t.Run("New", scheduler_New)
		t.Run(schedulerTest(scheduler_Actions))
		t.Run(schedulerTest(scheduler_CancelAll))
		t.Run(schedulerTest(scheduler_Schedule))
		t.Run(schedulerTest(scheduler_String))
		t.Run(schedulerTest(scheduler_executesWithDelay))
		t.Run(schedulerTest(scheduler_executesWithoutDelay))
	})
}

func scheduler_New(t *testing.T) {
	s := NewScheduler()
	if s == nil {
		t.Fatal("nil scheduler")
	}
}

func scheduler_Actions(t *testing.T, s *Scheduler) {
	foo := func() {}
	s.Schedule("foo1", time.Hour, foo)
	s.Schedule("foo2", time.Hour, foo)
	s.Schedule("foo3", time.Hour, foo)
	s.Schedule("foo4", time.Hour, foo)
	s.Schedule("foo5", time.Hour, foo)

	actions := s.Actions()

	if len(actions) != 5 {
		t.Fatalf("unexpected actions: %#v", actions)
	}
}

func scheduler_CancelAll(t *testing.T, s *Scheduler) {
	foo := func() {}
	s.Schedule("foo1", time.Hour, foo)
	s.Schedule("foo2", time.Hour, foo)
	s.Schedule("foo3", time.Hour, foo)
	s.Schedule("foo4", time.Hour, foo)
	s.Schedule("foo5", time.Hour, foo)

	s.CancelAll()

	actions := s.Actions()
	if len(actions) != 0 {
		t.Fatalf("unexpected actions: %#v", actions)
	}
}

func scheduler_Schedule(t *testing.T, s *Scheduler) {
	foo := func() {}
	s.Schedule("foo1", time.Hour, foo)
	s.Schedule("foo2", time.Hour, foo)
	s.Schedule("foo3", time.Hour, foo)
	s.Schedule("foo4", time.Hour, foo)
	s.Schedule("foo5", time.Hour, foo)

	actions := s.Actions()

	if len(actions) != 5 {
		t.Fatalf("unexpected actions: %#v", actions)
	}
}

func scheduler_String(t *testing.T, s *Scheduler) {
	foo := func() {}
	s.Schedule("foo1", time.Hour, foo)
	s.Schedule("foo2", time.Hour, foo)
	s.Schedule("foo3", time.Hour, foo)
	s.Schedule("foo4", time.Hour, foo)
	s.Schedule("foo5", time.Hour, foo)

	out := s.String()

	for _, name := range []string{"foo1", "foo2", "foo3", "foo4", "foo5"} {
		if !strings.Contains(out, name) {
			t.Errorf("%s not in %s", name, out)
		}
	}
}

func scheduler_executesWithoutDelay(t *testing.T, s *Scheduler) {
	var count = 0
	foo := func() { count++ }
	s.Schedule("foo", 0, foo)

	if count != 1 {
		t.Errorf("unexpected count: %d", count)
	}
}

func scheduler_executesWithDelay(t *testing.T, s *Scheduler) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	var count = 0
	foo := func() {
		defer wg.Done()
		count++
	}
	s.Schedule("foo", time.Millisecond, foo)

	wg.Wait()

	if count != 1 {
		t.Errorf("unexpected count: %d", count)
	}
}
