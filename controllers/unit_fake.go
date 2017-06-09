package controllers

import (
	"log"
	"time"

	"github.com/explodes/greenhouse-pi/stats"
)

type fakeUnit struct {
	statType stats.StatType
	on       bool
	storage  stats.Storage
}

func NewFakeUnit(statType stats.StatType, storage stats.Storage) Unit {
	return &fakeUnit{
		statType: statType,
		on:       false,
		storage:  storage,
	}
}

func (u *fakeUnit) Name() string {
	return string(u.statType)
}

func (u *fakeUnit) On() error {
	log.Printf("%s on", u.statType)
	u.on = true
	if err := u.storage.Record(stats.Stat{StatType: u.statType, When: time.Now(), Value: 1}); err != nil {
		return err
	}
	return nil
}

func (u *fakeUnit) Off() error {
	log.Printf("%s off", u.statType)
	u.on = false
	if err := u.storage.Record(stats.Stat{StatType: u.statType, When: time.Now(), Value: 0}); err != nil {
		return err
	}
	return nil
}

func (u *fakeUnit) Status() (UnitStatus, error) {
	if u.on {
		return UnitStatusOn, nil
	}
	return UnitStatusOff, nil
}

func (u *fakeUnit) Close() error {
	return nil
}
