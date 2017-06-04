package stats

import "time"

type Stat struct {
	StatType StatType
	When     time.Time
	Value    float64
}

type Storage interface {
	Record(stat Stat) error
	Fetch(statType StatType, start, end time.Time) ([]Stat, error)
}
