package mysql

import (
	"time"

	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/jobqueue/logger"
)

// incomingJob : implements the following interfaces
// - jobqueue.IncomingJob
// - jobqueue.Job
// - logger.LoggableJob
type incomingJob struct {
	jobqueue.IncomingJob
	id uint64
}

func (j *incomingJob) ID() uint64 {
	return j.id
}

func (j *incomingJob) FailCount() uint {
	return 0
}

func (j *incomingJob) Status() string {
	return "claimed"
}

func (j *incomingJob) CreatedAt() uint64 {
	return uint64(time.Now().UnixNano() / int64(time.Millisecond))
}

func (j *incomingJob) NextDelay() uint64 {
	return j.IncomingJob.NextDelay()
}

func (j *incomingJob) NextTry() uint64 {
	nowMillisecond := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	return nowMillisecond + j.NextDelay()
}

func (j *incomingJob) ToLoggable() logger.LoggableJob {
	return j
}

// job : implements the following interfaces
// - jobqueue.Job
// - logger.LoggableJob
type job struct {
	id         uint64
	category   string
	url        string
	payload    string
	status     string
	createdAt  uint64 // milliseconds
	nextTry    uint64 // milliseconds
	timeout    uint   // seconds
	retryDelay uint   // seconds
	retryCount uint
	failCount  uint
}

func (j *job) ID() uint64 {
	return j.id
}

func (j *job) Category() string {
	return j.category
}

func (j *job) URL() string {
	return j.url
}

func (j *job) Payload() string {
	return j.payload
}

func (j *job) NextTry() uint64 {
	return j.nextTry
}

func (j *job) RetryCount() uint {
	return j.retryCount
}

func (j *job) RetryDelay() uint {
	return j.retryDelay
}

func (j *job) FailCount() uint {
	return j.failCount
}

func (j *job) Timeout() uint {
	return j.timeout
}

func (j *job) Status() string {
	return j.status
}

func (j *job) CreatedAt() uint64 {
	return j.createdAt
}

func (j *job) ToLoggable() logger.LoggableJob {
	return j
}
