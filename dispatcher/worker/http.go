package worker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fireworq/fireworq/config"
	"github.com/fireworq/fireworq/jobqueue"

	"github.com/rs/zerolog"
)

var defaultUserAgent string

// HTTPInit initializes global parameters of HTTP workers by
// configuration values.
//
// Configuration keys prefixed by "dispatch_" are considered.
func HTTPInit() {
	transport := http.DefaultTransport.(*http.Transport)
	transport.MaxIdleConns = 0

	b, err := strconv.ParseBool(config.Get("dispatch_keep_alive"))
	if err != nil {
		b, _ = strconv.ParseBool(config.GetDefault("dispatch_keep_alive"))
	}
	transport.DisableKeepAlives = !b

	v, err := strconv.ParseInt(config.Get("dispatch_max_conns_per_host"), 10, 32)
	if err != nil {
		v, _ = strconv.ParseInt(config.GetDefault("dispatch_max_conns_per_host"), 10, 32)
	}
	transport.MaxIdleConnsPerHost = int(v)

	defaultUserAgent = config.Get("dispatch_user_agent")
}

// HTTPWorker is a worker which handles a job as an HTTP POST request
// to the URL specified by the job.
type HTTPWorker struct {
	UserAgent string
	Logger    *zerolog.Logger
}

// NewWorker creates a new HTTP worker instance which inherits the
// configurations of the current one.
func (worker *HTTPWorker) NewWorker() Worker {
	w := *worker

	if w.UserAgent == "" {
		w.UserAgent = defaultUserAgent
	}

	if w.Logger == nil {
		logger := zerolog.Nop()
		w.Logger = &logger
	}

	return &w
}

// Work makes a POST request to job.URL and returns the result.
func (worker *HTTPWorker) Work(job jobqueue.Job) *jobqueue.Result {
	client := &http.Client{
		Timeout: time.Duration(job.Timeout()) * time.Second,
	}
	req, err := http.NewRequest(
		"POST",
		job.URL(),
		strings.NewReader(job.Payload()),
	)
	if err != nil {
		return &jobqueue.Result{
			Status:  jobqueue.ResultStatusInternalFailure,
			Message: fmt.Sprintf("Cannot create http request: %v", err),
		}
	}
	req.Header.Add("Content-Type", "application/json")

	userAgent := worker.UserAgent
	if userAgent == "" {
		userAgent = defaultUserAgent
	}
	req.Header.Add("User-Agent", worker.UserAgent)

	resp, err := client.Do(req)

	worker.Logger.Debug().
		Str("action", "dispatch").
		Str("worker", "HTTPWorker").
		Str("url", job.URL()).
		Str("payload", job.Payload()).
		Msg("Dispatched via HTTP")

	if err != nil {
		return &jobqueue.Result{
			Status:  jobqueue.ResultStatusInternalFailure,
			Message: fmt.Sprintf("Request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &jobqueue.Result{
			Status:  jobqueue.ResultStatusFailure,
			Code:    resp.StatusCode,
			Message: fmt.Sprintf("Cannot read body: %v", err),
		}
	}

	var rslt jobqueue.Result
	err = json.Unmarshal(body, &rslt)
	if err != nil {
		return &jobqueue.Result{
			Status: jobqueue.ResultStatusFailure,
			Code:   resp.StatusCode,
			Message: fmt.Sprintf(
				"Cannot parse body as JSON: %v\nOriginal response body:\n%s",
				err,
				string(body),
			),
		}
	}

	if !rslt.IsValid() {
		return &jobqueue.Result{
			Status:  jobqueue.ResultStatusFailure,
			Code:    resp.StatusCode,
			Message: fmt.Sprintf("Invalid result status: %s\nOriginal response body:\n%s", rslt.Status, string(body)),
		}
	}

	rslt.Code = resp.StatusCode
	return &rslt
}
