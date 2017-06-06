package logging

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

type LogLevel uint8

type Logger interface {
	// Log records a message at a given log level
	Log(level LogLevel, fmt string, args ...interface{})
}

func (level LogLevel) String() string {
	switch level {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warning"
	case LogLevelError:
		return "error"
	default:
		return "unknown"
	}
}
