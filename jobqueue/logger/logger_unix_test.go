//go:build linux || darwin

package logger

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/fireworq/fireworq/config"
)

func TestLogOutput(t *testing.T) {
	fname := "logger-test.log"
	config.Locally("queue_log", fname, func() {
		Init()
		Writer.Write([]byte("foo\n"))
		defer os.Remove(fname)

		buf, err := ioutil.ReadFile(fname)
		if err != nil {
			t.Error(err)
		}
		if string(buf) != "foo\n" {
			t.Error("It must be able to write a log entry to a file")
		}
	})
}
