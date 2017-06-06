package controllers

import (
	"fmt"
	"time"

	"github.com/explodes/greenhouse-pi/logging"
	"github.com/explodes/greenhouse-pi/stats"
)

// Controller manages the timing of a Unit
type Controller struct {
	Unit      Unit
	storage   stats.Storage
	scheduler *Scheduler

	// isOn is whether or not the water Unit is known to be on
	isOn bool
}

func NewController(unit Unit, storage stats.Storage, scheduler *Scheduler) (*Controller, error) {
	isOn, err := unit.Status()
	if err != nil {
		return nil, fmt.Errorf("error creating controller for Unit %s: %v", unit.Name(), err)
	}
	wc := &Controller{
		scheduler: scheduler,
		Unit:      unit,
		storage:   storage,
		isOn:      isOn == UnitStatusOn,
	}
	return wc, nil
}

func (wc *Controller) TurnUnitOn(delay time.Duration, duration time.Duration) {
	wc.scheduler.Schedule(fmt.Sprintf("turn on %s", wc.Unit.Name()), delay, func() {
		wc.turnUnitOnNow()
	})
	wc.scheduler.Schedule(fmt.Sprintf("turn off %s", wc.Unit.Name()), delay+duration, func() {
		wc.turnUnitOffNow()
	})
}

func (wc *Controller) turnUnitOnNow() {
	if !wc.isOn {
		if err := wc.Unit.On(); err != nil {
			go wc.storage.Log(logging.LogLevelError, "error turning on water: %v", err)
		} else {
			go wc.storage.Log(logging.LogLevelInfo, "water was turned on")
			wc.isOn = true
		}
	}
}

func (wc *Controller) TurnUnitOff() {
	wc.turnUnitOffNow()
}

func (wc *Controller) turnUnitOffNow() {
	if wc.isOn {
		if err := wc.Unit.Off(); err != nil {
			go wc.storage.Log(logging.LogLevelError, "error turning off water: %v", err)
		} else {
			go wc.storage.Log(logging.LogLevelInfo, "water was turned off")
			wc.isOn = false
		}
	}
}
