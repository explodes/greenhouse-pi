package controllers

import (
	"context"
	"sync"
	"time"
)

type toggler interface {
	toggle(bool)
}

type Scheduler struct {
	taskLock *sync.Mutex

	// actions is a map of actions to whether or not they are enabled
	// so the pending actions can be queried
	actions map[*Action]bool
}

type Action struct {
	Name    string
	Created time.Time
	Start   time.Time

	Perform func()
	Cancel  func()
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		actions: make(map[*Action]bool),
	}
}

func (s *Scheduler) addAction(action *Action) {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	s.actions[action] = true
}

func (s *Scheduler) removeAction(action *Action) {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	delete(s.actions, action)
}

func (s *Scheduler) Schedule(name string, delay time.Duration, action func()) *Action {

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-time.After(delay):
			action()
		case <-ctx.Done():
			break
		}
	}()

	now := time.Now()
	start := now.Add(delay)

	a := &Action{
		Name:    name,
		Created: now,
		Start:   start,
		Perform: action,
	}
	a.Cancel = func() {
		cancel()
		s.removeAction(a)
	}
	s.addAction(a)

	return a
}

func (s *Scheduler) Actions() []*Action {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	actions := make([]*Action, 0, len(s.actions))
	for action := range s.actions {
		actions = append(actions, action)
	}
	return actions
}
