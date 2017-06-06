package stats

import (
	"errors"
	"time"

	"github.com/explodes/greenhouse-pi/logging"
)

type LogLevel uint8

type Log struct {
	Level   logging.LogLevel
	When    time.Time
	Message string
}

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

	// Log records a message at a given log level
	Log(level logging.LogLevel, fmt string, args ...interface{})

	// Logs retrieves logs for a given time frame with a given minimum log level
	Logs(level logging.LogLevel, start, end time.Time) ([]Log, error)
}
