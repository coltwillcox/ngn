package logz

import (
	"fmt"
	"strings"
)

type LogLevel int

const (
	LogTrace LogLevel = iota
	LogDebug
	LogInfo
	LogWarn
	LogErr
	NoLogs
)

func (level LogLevel) Eq(other LogLevel) bool {
	return int(level) == int(other)
}
func (level LogLevel) Lt(other LogLevel) bool {
	return int(level) < int(other)
}
func (level LogLevel) Le(other LogLevel) bool {
	return int(level) <= int(other)
}
func (level LogLevel) Gt(other LogLevel) bool {
	return int(level) > int(other)
}
func (level LogLevel) Ge(other LogLevel) bool {
	return int(level) >= int(other)
}

func (level LogLevel) String() string {
	switch level {
	case LogTrace:
		return "trace"
	case LogDebug:
		return "debug"
	case LogInfo:
		return "info"
	case LogWarn:
		return "warning"
	case LogErr:
		return "error"
	default:
		return "no_logs"
	}
}

// ToLevel may be used to convert a string to the corresponding
// debug level, for example when parsing from environmento or command
// line.
func ToLevel(level string) (LogLevel, error) {
	switch strings.ToLower(level) {
	case "trace":
		return LogTrace, nil
	case "debug":
		return LogDebug, nil
	case "info":
		return LogInfo, nil
	case "warning", "warn":
		return LogWarn, nil
	case "error", "err":
		return LogErr, nil
	case "silence", "no":
		return NoLogs, nil
	default:
		return NoLogs, fmt.Errorf("logs: unrecognized level '%s'", level)
	}
}

// Logger is the basic interface. All implementations provided by
// git.sr.ht/~blallo/logz must follow it.
type Logger interface {
	Log(level LogLevel, data map[string]interface{})
	Trace(data map[string]interface{})
	Debug(data map[string]interface{})
	Info(data map[string]interface{})
	Warn(data map[string]interface{})
	Err(data map[string]interface{})
	SetLevel(LogLevel) Logger
}
