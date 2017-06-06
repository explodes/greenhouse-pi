package controllers

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"
)

// Scheduler schedules actions and provides a way to view what is queued up
type Scheduler struct {
	taskLock *sync.Mutex

	// actions is a set of pending actions
	actions map[*Action]bool
}

// Action is a function that is to be called some time in the future
type Action struct {
	// Name is the name of the action being performed
	Name string
	// Created is when this action was queued
	Created time.Time
	// Start is when this action is scheduled to start
	Start time.Time

	// Perform will perform this action
	Perform func()
	// Cancel will cancel this action
	Cancel func()
}

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		taskLock: &sync.Mutex{},
		actions:  make(map[*Action]bool),
	}
}

// addAction puts an action in the set of pending actions
func (s *Scheduler) addAction(action *Action) {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	s.actions[action] = true
}

// removeAction removes an action from the set of pending actions
func (s *Scheduler) removeAction(action *Action) {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	delete(s.actions, action)
}

// Schedule executes an action in the future and returns that action.
// If there is no delay, the action is executed immediately and no Action is returned.
func (s *Scheduler) Schedule(name string, delay time.Duration, action func()) *Action {

	if delay <= 0 {
		action()
		return nil
	}

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

// Actions returns a list of pending actions
func (s *Scheduler) Actions() []*Action {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	actions := make([]*Action, 0, len(s.actions))
	for action := range s.actions {
		actions = append(actions, action)
	}
	return actions
}

// CancelAll cancels all pending actions in the scheduler
func (s *Scheduler) CancelAll() {
	for action := range s.actions {
		action.Cancel()
	}
}

func (s *Scheduler) String() string {
	buf := bytes.NewBuffer(make([]byte, 0, 128))

	buf.WriteString("Scheduler:{")

	actions := s.Actions()
	for _, action := range actions {

		buf.WriteString(fmt.Sprintf("{%s:%s}", action.Name, action.Start))
	}

	buf.WriteRune('}')

	return string(buf.Bytes())
}
