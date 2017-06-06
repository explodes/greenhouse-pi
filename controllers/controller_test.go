package controllers

import (
	"sync"
	"testing"
	"time"

	"github.com/explodes/greenhouse-pi/stats"
)

type TestUnit struct {
	status bool
	wg     *sync.WaitGroup
}

func (u *TestUnit) Name() string {
	return "testunit"
}

func (u *TestUnit) On() error {
	defer u.wg.Done()
	u.status = true
	return nil
}

func (u *TestUnit) Off() error {
	defer u.wg.Done()
	u.status = false
	return nil
}

func (u *TestUnit) Status() (UnitStatus, error) {
	if u.status {
		return UnitStatusOn, nil
	}
	return UnitStatusOff, nil
}

func controllerTest(f func(t *testing.T, c *Controller, testUnit *TestUnit)) (string, func(*testing.T)) {
	return testFunctionName(f), func(t *testing.T) {
		t.Parallel()

		unit := &TestUnit{wg: &sync.WaitGroup{}}
		storage := stats.NewFakeStatsStorage(40)
		scheduler := NewScheduler()

		c, err := NewController(unit, storage, scheduler)
		if err != nil {
			t.Fatal(err)
		}
		f(t, c, unit)

		scheduler.CancelAll()
	}
}

func TestController(t *testing.T) {
	t.Run("Scheduler", func(t *testing.T) {
		t.Parallel()
		t.Run("New", controller_New)
		t.Run(controllerTest(controller_TurnUnitOff))
		t.Run(controllerTest(controller_TurnUnitOn))
	})
}

func controller_New(t *testing.T) {
	unit := NewFakeUnit("test")
	storage := stats.NewFakeStatsStorage(40)
	scheduler := NewScheduler()

	c, err := NewController(unit, storage, scheduler)
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("nil controller")
	}
}

func controller_TurnUnitOff(t *testing.T, c *Controller, testUnit *TestUnit) {
	testUnit.wg.Add(1)

	c.turnUnitOffNow()

	status, err := c.Unit.Status()
	if err != nil {
		t.Fatal(err)
	}
	if status != UnitStatusOff {
		t.Error("controller did not turn off Unit")
	}
}

func controller_TurnUnitOn(t *testing.T, c *Controller, testUnit *TestUnit) {
	testUnit.wg.Add(2)

	c.TurnUnitOn(time.Millisecond, 10*time.Millisecond)

	testUnit.wg.Wait()

	if testUnit.status {
		t.Error("test Unit did not turn off")
	}

}
