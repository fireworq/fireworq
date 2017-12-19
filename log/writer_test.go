package log

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	writer := New(os.Stdout)
	writer.Reopen()
}
