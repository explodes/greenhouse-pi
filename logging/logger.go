package logging

import "time"

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type LogEntry struct {
	Level   Level
	When    time.Time
	Message string
}

type Level uint8

type Logger interface {
	// Log records a message at a given log level
	Log(level Level, fmt string, args ...interface{}) (LogEntry, error)
}

func (level Level) String() string {
	switch level {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warning"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}
