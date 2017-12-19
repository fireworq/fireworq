// +build linux darwin

package log

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestOpenFile(t *testing.T) {
	fname := "log-writer-open-file.log"
	writer, err := OpenFile(fname)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(fname)

	content := "This is a test log entry\n"
	writer.Write([]byte(content))

	{
		buf, err := ioutil.ReadFile(fname)
		if err != nil {
			t.Error(err)
		}
		if string(buf) != content {
			t.Error("File writer should write a log entry to a file")
		}
	}

	renamed := "log-writer-renamed-file.log"
	if err := os.Rename(fname, renamed); err != nil {
		t.Error(err)
	}
	defer os.Remove(renamed)

	writer.Reopen()

	another := "This is another test log entry\n"
	writer.Write([]byte(another))

	{
		buf, err := ioutil.ReadFile(fname)
		if err != nil {
			t.Error(err)
		}
		if string(buf) != another {
			t.Error("File writer should write a log entry to a file")
		}
	}

	{
		buf, err := ioutil.ReadFile(renamed)
		if err != nil {
			t.Error(err)
		}
		if string(buf) != content {
			t.Error("File writer should not touch a moved file after Reopen()")
		}
	}
}
