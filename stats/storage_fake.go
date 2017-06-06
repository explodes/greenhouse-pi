package stats

import (
	"sync"
	"time"
)

type fakeStatsStorage struct {
	mu      *sync.RWMutex
	storage map[StatType][]Stat
	limit   int
}

func NewFakeStatsStorage(limit int) Storage {
	return &fakeStatsStorage{
		mu:      &sync.RWMutex{},
		storage: make(map[StatType][]Stat),
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
