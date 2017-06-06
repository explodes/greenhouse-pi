package stats

import (
	"errors"
	"time"
)

type Stat struct {
	StatType StatType
	When     time.Time
	Value    float64
}

var (
	// ErrNoStats indicates that there are no
	// statistics found of a particular type
	ErrNoStats = errors.New("no stats found")
)

// Storage is the database interface to
// record and retrieve statistics
type Storage interface {
	// Record puts a Stat record in the Storage
	Record(stat Stat) error

	// Fetch retrieves a list of a particular Stat for a given time frame
	Fetch(statType StatType, start, end time.Time) ([]Stat, error)

	// Latest fetches the latest Stat of a particular
	// type from the Storage.  If there are no statistics
	// of that type recorded, it should return ErrNoStats
	Latest(statType StatType) (Stat, error)
}
