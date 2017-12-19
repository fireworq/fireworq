package log

import (
	"strings"

	"github.com/rs/zerolog"
)

// ParseLevel parses a string which represents a log level and returns
// a zerolog.Level.
func ParseLevel(level string, defaultLevel zerolog.Level) zerolog.Level {
	l := defaultLevel
	switch strings.ToLower(level) {
	case "0", "debug":
		l = zerolog.DebugLevel
	case "1", "info":
		l = zerolog.InfoLevel
	case "2", "warn":
		l = zerolog.WarnLevel
	case "3", "error":
		l = zerolog.ErrorLevel
	case "4", "fatal":
		l = zerolog.FatalLevel
	}
	return l
}
