package jobqueue

import (
	"time"

	"github.com/fireworq/fireworq/jobqueue/logger"
)

// IncomingJob is an interface of incoming jobs.
type IncomingJob interface {
	Category() string
	URL() string
	Payload() string

	NextDelay() uint64 // milliseconds
	Timeout() uint     // seconds
	RetryDelay() uint  // seconds
	RetryCount() uint
}

// Job is an interface of jobs.
type Job interface {
	URL() string
	Payload() string
	Timeout() uint

	RetryCount() uint
	RetryDelay() uint
	FailCount() uint

	ToLoggable() logger.LoggableJob
}

// completedJob : implements the following interfaces
// - Job
// - logger.LoggableJob
type completedJob struct {
	Job
	failed uint
}

func (j *completedJob) FailCount() uint {
	return j.Job.FailCount() + j.failed
}

func (j *completedJob) Status() string {
	return "completed"
}

func (j *completedJob) ToLoggable() logger.LoggableJob {
	return &loggableCompletedJob{
		j.Job.ToLoggable(),
		j.Status(),
		j.FailCount(),
	}
}

func (j *completedJob) canRetry() bool {
	return j.Job.RetryCount() > 0
}

type loggableCompletedJob struct {
	logger.LoggableJob
	status    string
	failCount uint
}

func (j *loggableCompletedJob) Status() string {
	return j.status
}

func (j *loggableCompletedJob) FailCount() uint {
	return j.failCount
}

// nextJob : implements the following interfaces
// - NextInfo
type nextJob struct {
	job Job
}

func (j *nextJob) NextDelay() uint64 {
	return uint64(time.Duration(j.job.RetryDelay()) * time.Second / time.Millisecond)
}

func (j *nextJob) RetryCount() uint {
	return j.job.RetryCount() - 1
}

func (j *nextJob) FailCount() uint {
	return j.job.FailCount()
}

// NextInfo describes information of a retry.
type NextInfo interface {
	NextDelay() uint64
	RetryCount() uint
	FailCount() uint
}
