package worker

import (
	"github.com/fireworq/fireworq/jobqueue"
)

// Worker is an interface of a worker which handles a dispatched job.
type Worker interface {
	Work(job jobqueue.Job) *jobqueue.Result
}

// Config is an interface of a builder of Worker.
type Config interface {
	NewWorker() Worker
}
