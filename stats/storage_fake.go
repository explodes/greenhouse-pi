package stats

import (
	"fmt"
	"sync"
	"time"

	"github.com/explodes/greenhouse-pi/logging"
	"log"
)

type fakeStatsStorage struct {
	mu      *sync.RWMutex
	storage map[StatType][]Stat
	logs    []Log
	limit   int
}

func NewFakeStatsStorage(limit int) Storage {
	return &fakeStatsStorage{
		mu:      &sync.RWMutex{},
		storage: make(map[StatType][]Stat),
		logs:    make([]Log, limit),
		limit:   limit,
	}
}

func (ss *fakeStatsStorage) Record(stat Stat) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	list, ok := ss.storage[stat.StatType]
	if !ok {
		list = make([]Stat, 0, ss.limit)
	}

	if len(list) > ss.limit {
		list = list[1:]
	}

	list = append(list, stat)

	ss.storage[stat.StatType] = list

	return nil
}

func between(when, start, end time.Time) bool {
	return (when.Equal(start) || when.After(start)) && (when.Equal(end) || when.Before(end))
}

func (ss *fakeStatsStorage) Fetch(statType StatType, start, end time.Time) ([]Stat, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	list, ok := ss.storage[statType]
	if !ok {
		return []Stat{}, nil
	}

	filtered := make([]Stat, 0, ss.limit)
	for _, stat := range list {
		if stat.StatType == statType && between(stat.When, start, end) {
			filtered = append(filtered, stat)
		}
	}

	return filtered, nil
}

func (ss *fakeStatsStorage) Latest(statType StatType) (Stat, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	list, ok := ss.storage[statType]
	if !ok || len(list) == 0 {
		return Stat{}, ErrNoStats
	}

	latest := list[0]
	for _, stat := range list[1:] {
		if stat.StatType == statType && stat.When.After(latest.When) {
			latest = stat
		}
	}

	return latest, nil
}

func (ss *fakeStatsStorage) Log(level logging.LogLevel, format string, args ...interface{}) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	msg := fmt.Sprintf(format, args...)

	if len(ss.logs) > ss.limit {
		ss.logs = ss.logs[1:]
	}

	ss.logs = append(ss.logs, Log{
		When:    time.Now(),
		Level:   level,
		Message: msg,
	})

	log.Printf("%s: %s", level, msg)
}

func (ss *fakeStatsStorage) Logs(level logging.LogLevel, start, end time.Time) ([]Log, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	filtered := make([]Log, 0, ss.limit)
	for _, entry := range ss.logs {
		if entry.Level >= level && between(entry.When, start, end) {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}
