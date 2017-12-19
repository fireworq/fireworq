package log

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestParseLevel(t *testing.T) {
	if ParseLevel("0", zerolog.FatalLevel) != zerolog.DebugLevel {
		t.Error("'0' should be debug level")
	}
	if ParseLevel("debug", zerolog.FatalLevel) != zerolog.DebugLevel {
		t.Error("'debug' should be debug level")
	}
	if ParseLevel("Debug", zerolog.FatalLevel) != zerolog.DebugLevel {
		t.Error("'Debug' should be debug level")
	}
	if ParseLevel("DEBUG", zerolog.FatalLevel) != zerolog.DebugLevel {
		t.Error("'DEBUG' should be debug level")
	}

	if ParseLevel("1", zerolog.FatalLevel) != zerolog.InfoLevel {
		t.Error("'1' should be info level")
	}
	if ParseLevel("info", zerolog.FatalLevel) != zerolog.InfoLevel {
		t.Error("'info' should be info level")
	}
	if ParseLevel("Info", zerolog.FatalLevel) != zerolog.InfoLevel {
		t.Error("'Info' should be info level")
	}
	if ParseLevel("INFO", zerolog.FatalLevel) != zerolog.InfoLevel {
		t.Error("'INFO' should be info level")
	}

	if ParseLevel("2", zerolog.FatalLevel) != zerolog.WarnLevel {
		t.Error("'2' should be warn level")
	}
	if ParseLevel("warn", zerolog.FatalLevel) != zerolog.WarnLevel {
		t.Error("'warn' should be warn level")
	}
	if ParseLevel("Warn", zerolog.FatalLevel) != zerolog.WarnLevel {
		t.Error("'Warn' should be warn level")
	}
	if ParseLevel("WARN", zerolog.FatalLevel) != zerolog.WarnLevel {
		t.Error("'WARN' should be warn level")
	}

	if ParseLevel("3", zerolog.FatalLevel) != zerolog.ErrorLevel {
		t.Error("'3' should be error level")
	}
	if ParseLevel("error", zerolog.FatalLevel) != zerolog.ErrorLevel {
		t.Error("'error' should be error level")
	}
	if ParseLevel("Error", zerolog.FatalLevel) != zerolog.ErrorLevel {
		t.Error("'Error' should be error level")
	}
	if ParseLevel("ERROR", zerolog.FatalLevel) != zerolog.ErrorLevel {
		t.Error("'ERROR' should be error level")
	}

	if ParseLevel("4", zerolog.WarnLevel) != zerolog.FatalLevel {
		t.Error("'4' should be fatal level")
	}
	if ParseLevel("fatal", zerolog.WarnLevel) != zerolog.FatalLevel {
		t.Error("'fatal' should be fatal level")
	}
	if ParseLevel("Fatal", zerolog.WarnLevel) != zerolog.FatalLevel {
		t.Error("'Fatal' should be fatal level")
	}
	if ParseLevel("FATAL", zerolog.WarnLevel) != zerolog.FatalLevel {
		t.Error("'FATAL' should be fatal level")
	}
}
