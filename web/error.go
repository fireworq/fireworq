package web

import (
	"fmt"
	"net/http"
)

type clientError interface {
	error
	clientError() string
	httpStatus() int
}

type simpleClientError int

func (e simpleClientError) httpStatus() int {
	return int(e)
}

func (e simpleClientError) Error() string {
	status := e.httpStatus()
	return fmt.Sprintf("%d %s", status, http.StatusText(status))
}

func (e simpleClientError) clientError() string {
	return e.Error()
}

func (e simpleClientError) WithDetail(detail string) *detailedClientError {
	return &detailedClientError{
		status:  int(e),
		message: detail,
	}
}

type detailedClientError struct {
	status  int
	message string
}

func (e *detailedClientError) httpStatus() int {
	return e.status
}

func (e *detailedClientError) Error() string {
	status := e.httpStatus()
	return fmt.Sprintf("%d %s\n\n%s", status, http.StatusText(status), e.message)
}

func (e *detailedClientError) clientError() string {
	return e.Error()
}

type serverError interface {
	error
	serverError() string
	httpStatus() int
}

type simpleServerError int

func (e simpleServerError) httpStatus() int {
	return int(e)
}

func (e simpleServerError) Error() string {
	status := e.httpStatus()
	return fmt.Sprintf("%d %s", status, http.StatusText(status))
}

func (e simpleServerError) serverError() string {
	return e.Error()
}

func (e simpleServerError) WithDetail(detail string) *detailedServerError {
	return &detailedServerError{
		status:  int(e),
		message: detail,
	}
}

type detailedServerError struct {
	status  int
	message string
}

func (e *detailedServerError) httpStatus() int {
	return e.status
}

func (e *detailedServerError) Error() string {
	status := e.httpStatus()
	return fmt.Sprintf("%d %s\n\n%s", status, http.StatusText(status), e.message)
}

func (e *detailedServerError) serverError() string {
	return e.Error()
}

const (
	errMethodNotAllowed    = simpleClientError(http.StatusMethodNotAllowed)
	errNotFound            = simpleClientError(http.StatusNotFound)
	errBadRequest          = simpleClientError(http.StatusBadRequest)
	errNotImplemented      = simpleServerError(http.StatusNotImplemented)
	errInternalServerError = simpleServerError(http.StatusInternalServerError)
)
