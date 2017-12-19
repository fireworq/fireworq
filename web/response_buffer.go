package web

import (
	"bytes"
	"net/http"
	"strconv"
)

type responseBuffer struct {
	buf    bytes.Buffer
	status int
	header http.Header
}

func (rb *responseBuffer) Write(buf []byte) (int, error) {
	return rb.buf.Write(buf)
}

func (rb *responseBuffer) WriteHeader(status int) {
	rb.status = status
}

func (rb *responseBuffer) Header() http.Header {
	if rb.header == nil {
		rb.header = make(http.Header)
	}
	return rb.header
}

func (rb *responseBuffer) WriteTo(w http.ResponseWriter) error {
	for k, v := range rb.header {
		w.Header()[k] = v
	}
	if rb.buf.Len() > 0 {
		w.Header().Set("Content-Length", strconv.Itoa(rb.buf.Len()))
	}
	if rb.status != 0 {
		w.WriteHeader(rb.status)
	}
	if rb.buf.Len() > 0 {
		if _, err := rb.buf.WriteTo(w); err != nil {
			return err
		}
	}
	return nil
}
