package controllers

import (
	"fmt"
	"sync"
	"time"

	"github.com/explodes/greenhouse-pi/stats"
)

// WaterUnit toggles the flow of water in the system
type WaterUnit interface {
	// On turns on the water unit
	On() error

	// Off turns off the water unit
	Off() error
}

// WaterController manages the timing of water
type WaterController struct {
	mu *sync.Mutex

	unit    WaterUnit
	storage stats.Storage

	// isOn is whether or not the water unit is known to be on
	isOn bool
}

func NewWaterController(unit WaterUnit, storage stats.Storage) (*WaterController, error) {
	isOn, err := getLatestOnValue(storage)
	if err != nil {
		return nil, fmt.Errorf("error creating water controller: %v", err)
	}

	wc := &WaterController{
		unit:    unit,
		storage: storage,
		isOn:    isOn,
	}
	return wc, nil
}

func getLatestOnValue(storage stats.Storage) (bool, error) {
	stat, err := storage.Latest(stats.StatTypeWater)
	if err == stats.ErrNoStats {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error fetching water stat: %v", err)
	}
	return stat.Value == float64(1), nil
}

func (wc *WaterController) TurnUnitOn(delay time.Duration, duration time.Duration) {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	if delay == 0 {
		wc.turnUnitOnNow()
		return
	}
}

func (wc *WaterController) turnUnitOnNow() {
	if !wc.isOn {
		if err := wc.unit.On(); err != nil {

			wc.isOn = true
		}
	}
}

func (wc *WaterController) TurnUnitOff() {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.turnUnitOffNow()
}

func (wc *WaterController) turnUnitOffNow() {
	if wc.isOn {
		wc.isOn = false
		wc.unit.Off()
	}
}
