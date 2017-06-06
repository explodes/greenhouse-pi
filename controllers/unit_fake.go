package controllers

import "log"

type fakeUnit struct {
	name   string
	status bool
}

func NewFakeUnit(name string) Unit {
	return &fakeUnit{
		name:   name,
		status: false,
	}
}

func (wu *fakeUnit) Name() string {
	return wu.name
}

func (wu *fakeUnit) On() error {
	log.Printf("%s on", wu.name)
	wu.status = true
	return nil
}

func (wu *fakeUnit) Off() error {
	log.Printf("%s off", wu.name)
	wu.status = false
	return nil
}

func (wu *fakeUnit) Status() (bool, error) {
	return wu.status, nil
}
