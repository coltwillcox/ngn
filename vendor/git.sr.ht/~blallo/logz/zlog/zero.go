package zlog

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"

	"git.sr.ht/~blallo/logz/interface"
)

// Logger is a git.sr.ht/~blallo/logz.Logger with a github.com/rs/zerolog backend.
// It might be used both as a json logger or a nicely colored console logger.
type Logger struct {
	level logz.LogLevel
	log   *zerolog.Logger
}

func (l *Logger) Log(level logz.LogLevel, data map[string]interface{}) {
	if level.Lt(l.level) {
		return
	}

	msg, ok := data["msg"].(string)
	if !ok {
		msg = ""
	}

	log := l.log.WithLevel(toZerologLevel(level))

	for k, v := range data {
		if k == "msg" {
			continue
		}

		log = log.Interface(k, v)
	}

	log.Msg(msg)
}

func (l *Logger) SetLevel(level logz.LogLevel) logz.Logger {
	l.level = level
	return l
}

func (l *Logger) Trace(data map[string]interface{}) { l.Log(logz.LogTrace, data) }
func (l *Logger) Debug(data map[string]interface{}) { l.Log(logz.LogDebug, data) }
func (l *Logger) Info(data map[string]interface{})  { l.Log(logz.LogInfo, data) }
func (l *Logger) Warn(data map[string]interface{})  { l.Log(logz.LogWarn, data) }
func (l *Logger) Err(data map[string]interface{})   { l.Log(logz.LogErr, data) }

// NewConsoleLogger initializes a logger with a github.com/rs/zerolog backend,
// configured to produce a nice colored output for usage in console applications
// (do not use this logger for production-grade high-log-throughput applications).
func NewConsoleLogger() logz.Logger {
	return withTime(newLogger(os.Stdout, true))
}

// NewConsoleLoggerStderr is the same as NewConsoleLogger, but writes on stderr
// instead of stdout.
func NewConsoleLoggerStderr() logz.Logger {
	return withTime(newLogger(os.Stderr, true))
}

// NewJSONLogger initializes a logger with a github.com/rs/zerolog backend,
// configured to produce json-formatted logs for high-log-throughput applications.
func NewJSONLogger() logz.Logger {
	return withTime(newLogger(os.Stdout, false))
}

func newLogger(w io.Writer, isConsole bool) *Logger {
	out := w
	if isConsole {
		out = zerolog.ConsoleWriter{Out: w}
	}

	log := zerolog.New(out)

	return &Logger{
		level: logz.LogInfo,
		log:   &log,
	}
}

func withTime(logger *Logger) *Logger {
	log := logger.log.With().Timestamp().Logger()

	return &Logger{
		level: logger.level,
		log:   &log,
	}
}

var _ logz.Logger = &Logger{}

func toZerologLevel(level logz.LogLevel) zerolog.Level {
	switch level {
	case logz.LogTrace:
		return zerolog.TraceLevel
	case logz.LogDebug:
		return zerolog.DebugLevel
	case logz.LogInfo:
		return zerolog.InfoLevel
	case logz.LogWarn:
		return zerolog.WarnLevel
	case logz.LogErr:
		return zerolog.ErrorLevel
	default:
		panic(fmt.Errorf("Unknown level %s", level))
	}
}
