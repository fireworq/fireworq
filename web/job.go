package web

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

func (app *Application) serveJob(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "POST" {
		return errMethodNotAllowed
	}

	vars := mux.Vars(req)

	var job IncomingJob
	decoder := json.NewDecoder(req.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&job); err != nil {
		return errBadRequest.WithDetail(err.Error())
	}
	if err := job.DecodePayload(); err != nil {
		return errBadRequest.WithDetail(err.Error())
	}
	if job.URLField == "" {
		return errBadRequest.WithDetail("Missing field: url")
	}
	job.CategoryField = vars["category"]

	r, err := app.Service.Push(&job)
	if err != nil {
		return err
	}

	result := PushResult{r.ID, r.QueueName, job}

	j, err := json.Marshal(&result)
	if err != nil {
		return err
	}

	writeJSON(w, j)
	return nil
}

// IncomingJob describes a job to be pushed in a queue.
type IncomingJob struct {
	CategoryField string          `json:"category"`
	URLField      string          `json:"url"`
	PayloadField  json.RawMessage `json:"payload"`
	payloadField  string

	RunAfterField   uint `json:"run_after"`   // seconds
	TimeoutField    uint `json:"timeout"`     // seconds
	RetryDelayField uint `json:"retry_delay"` // seconds
	MaxRetriesField uint `json:"max_retries"`
}

// PushResult describes a job pushed to a queue.
type PushResult struct {
	ID        uint64 `json:"id"`
	QueueName string `json:"queue_name"`
	IncomingJob
}

// Category returns the category of the job.
func (job *IncomingJob) Category() string {
	return job.CategoryField
}

// URL returns the URL of the job.
func (job *IncomingJob) URL() string {
	return job.URLField
}

// DecodePayload decodes PayloadField of the job.
//
// If PayloadField starts and ends with ", then it is decoded as a
// JSON string.  If PayloadField is "null" then it is decoded to an
// empty string.  Otherwise, the decoded value is the raw string of
// PayloadField.
//
// The decoded value can be retrieved by Payload() method.
func (job *IncomingJob) DecodePayload() error {
	payload := job.PayloadField
	if len(payload) > 0 && payload[0] == '"' && payload[len(payload)-1] == '"' {
		var buf string
		err := json.Unmarshal(payload, &buf)
		if err != nil {
			return errors.New("The payload seems to be a string but is broken")
		}
		job.payloadField = buf
	} else if len(payload) == 4 && payload[0] == 'n' && payload[1] == 'u' && payload[2] == 'l' && payload[3] == 'l' {
		job.payloadField = ""
	} else {
		job.payloadField = string(payload)
	}
	return nil
}

// Payload returns a decoded value of PayloadField.
func (job *IncomingJob) Payload() string {
	if job.payloadField == "" {
		job.DecodePayload()
	}
	return job.payloadField
}

// NextDelay returns the delay for a next try of the job.
func (job *IncomingJob) NextDelay() uint64 {
	return uint64(job.RunAfterField * 1000)
}

// RetryCount returns the max retries of the job.
func (job *IncomingJob) RetryCount() uint {
	return job.MaxRetriesField
}

// RetryDelay returns the delay for retries of the job.
func (job *IncomingJob) RetryDelay() uint {
	return job.RetryDelayField
}

// Timeout returns the timeout of the job.
func (job *IncomingJob) Timeout() uint {
	return job.TimeoutField
}
