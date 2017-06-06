package controllers

import "log"

type fakeUnit struct {
	name string
	on   bool
}

func NewFakeUnit(name string) Unit {
	return &fakeUnit{
		name: name,
		on:   false,
	}
}

func (u *fakeUnit) Name() string {
	return u.name
}

func (u *fakeUnit) On() error {
	log.Printf("%s on", u.name)
	u.on = true
	return nil
}

func (u *fakeUnit) Off() error {
	log.Printf("%s off", u.name)
	u.on = false
	return nil
}

func (u *fakeUnit) Status() (UnitStatus, error) {
	if u.on {
		return UnitStatusOn, nil
	}
	return UnitStatusOff, nil
}
