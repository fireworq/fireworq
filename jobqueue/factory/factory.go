package factory

import (
	"github.com/fireworq/fireworq/config"
	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/jobqueue/inmemory"
	"github.com/fireworq/fireworq/jobqueue/mysql"
	"github.com/fireworq/fireworq/model"

	"github.com/rs/zerolog/log"
)

// IncomingJob imitates IncomingJob in jobqueue package: factory
// package is intended to be used as a jobqueue package (by import
// jobqueue ".../fireworq/jobqueue/factory" since the only reason for
// having a separate package is to avoid cyclic import with a driver
// package such as mysql.
type IncomingJob = jobqueue.IncomingJob

// JobQueue imitates JobQueue in jobqueue package: factory package is
// intended to be used as a jobqueue package (by import jobqueue
// ".../fireworq/jobqueue/factory" since the only reason for having a
// separate package is to avoid cyclic import with a driver package
// such as mysql.
type JobQueue = jobqueue.JobQueue

// NewImpl creates a new jobqueue.Impl instance according to the value
// of "driver" configuration.
func NewImpl(q *model.Queue) jobqueue.Impl {
	log := log.With().Str("queue", q.Name).Logger()

	var impl jobqueue.Impl

	driver := config.Get("driver")
	if driver == "mysql" {
		log.Info().Msg("Select mysql as a driver for a job queue")
		impl = mysql.NewPrimaryBackup(q, mysql.Dsn())
	}
	if driver == "in-memory" {
		log.Info().Msg("Select in-memory as a driver for a job queue")
		impl = inmemory.New()
	}

	if impl == nil {
		log.Panic().Msgf("Unknown driver: %s", driver)
	}

	return impl
}

// Start creates and starts a new JobQueue instance whose
// implementation is decided by the value of "driver" configuration.
func Start(q *model.Queue) JobQueue {
	impl := NewImpl(q)
	return jobqueue.Start(q, impl)
}
