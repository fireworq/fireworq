package log

import (
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Writer is an io.Writer with Reopen() method.
type Writer interface {
	io.Writer
	Reopen() error
}

type writer struct {
	io.Writer
}

// New returns a new Writer which inherits w and does nothing on Reopen().
func New(w io.Writer) Writer {
	return &writer{w}
}

func (w *writer) Reopen() error {
	return nil
}

type fileWriter struct {
	file *os.File
	mu   sync.RWMutex
}

// OpenFile opens a file and returns a Writer which writes to the file
// and reopens the file on Reopen().
func OpenFile(name string) (Writer, error) {
	file, err := openFile(name)
	if err != nil {
		return nil, err
	}

	return &fileWriter{file: file}, nil
}

func openFile(name string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
}

func (w *fileWriter) Reopen() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	file, err := openFile(w.file.Name())
	if err != nil {
		return err
	}

	w.file.Close()
	w.file = file

	return nil
}

func (w *fileWriter) Write(p []byte) (int, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.file.Write(p)
}
