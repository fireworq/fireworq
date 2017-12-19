package web

import (
	"net/http"
)

type handler func(w http.ResponseWriter, req *http.Request) error

func (f handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if disableKeepAlives {
		w.Header().Set("Connection", "close")
	}

	var rb responseBuffer
	err := f(&rb, req)
	if err != nil {
		if ce, ok := err.(clientError); ok {
			http.Error(w, ce.clientError(), ce.httpStatus())
			return
		}
		if se, ok := err.(serverError); ok {
			http.Error(w, se.serverError(), se.httpStatus())
			return
		}

		se := errInternalServerError.WithDetail(err.Error())
		http.Error(w, se.serverError(), se.httpStatus())
	}
	rb.WriteTo(w)
}
