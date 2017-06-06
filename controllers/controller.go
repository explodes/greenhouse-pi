package controllers

import (
	"fmt"
	"time"

	"github.com/explodes/greenhouse-pi/logging"
	"github.com/explodes/greenhouse-pi/stats"
)

// Unit toggles the flow of water in the system
type Unit interface {
	// Name of this unit
	Name() string

	// On turns on the unit
	On() error

	// Off turns off the unit
	Off() error

	// Whether or not this unit is on
	Status() (bool, error)
}

// Controller manages the timing of a Unit
type Controller struct {
	unit      Unit
	storage   stats.Storage
	scheduler *Scheduler

	// isOn is whether or not the water unit is known to be on
	isOn bool
}

func NewController(unit Unit, storage stats.Storage, scheduler *Scheduler) (*Controller, error) {
	isOn, err := unit.Status()
	if err != nil {
		return nil, fmt.Errorf("error creating controller for unit %s: %v", unit.Name(), err)
	}
	wc := &Controller{
		scheduler: scheduler,
		unit:      unit,
		storage:   storage,
		isOn:      isOn,
	}
	return wc, nil
}

func (wc *Controller) TurnUnitOn(delay time.Duration, duration time.Duration) {
	wc.scheduler.Schedule(fmt.Sprintf("turn on %s", wc.unit.Name()), delay, func() {
		wc.turnUnitOnNow()
	})
	wc.scheduler.Schedule(fmt.Sprintf("turn off %s", wc.unit.Name()), delay+duration, func() {
		wc.turnUnitOffNow()
	})
}

func (wc *Controller) turnUnitOnNow() {
	if !wc.isOn {
		if err := wc.unit.On(); err != nil {
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
		if err := wc.unit.Off(); err != nil {
			go wc.storage.Log(logging.LogLevelError, "error turning off water: %v", err)
		} else {
			go wc.storage.Log(logging.LogLevelInfo, "water was turned off")
			wc.isOn = false
		}
	}
}
