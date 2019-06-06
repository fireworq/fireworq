package jobqueue

import (
	"github.com/fireworq/fireworq/jobqueue/logger"
	"github.com/fireworq/fireworq/model"

	"github.com/rs/zerolog/log"
)

// Impl is an interface of a job queue implementation.
type Impl interface {
	Start()
	Stop() <-chan struct{}
	Push(job IncomingJob) (Job, error)
	Pop(limit uint) ([]Job, error)
	Delete(job Job)
	Update(job Job, next NextInfo)
	IsActive() bool
}

// JobQueue is an interface of a job queue.
type JobQueue interface {
	Stop() <-chan struct{}
	Push(job IncomingJob) (uint64, error)
	Pop(limit uint) ([]Job, error)
	Complete(job Job, res *Result)

	Name() string

	IsActive() bool
	Node() (*Node, error)
	Stats() *Stats

	Inspector() (Inspector, bool)
	FailureLog() (FailureLog, bool)
}

// Start returns a job queue.
func Start(definition *model.Queue, q Impl) JobQueue {
	jq := &jobQueue{
		name:       definition.Name,
		maxWorkers: definition.MaxWorkers,
		impl:       q,
		stats:      newStats(),
	}
	q.Start()
	return jq
}

type jobQueue struct {
	name       string
	maxWorkers uint
	impl       Impl
	stats      *stats
}

func (q *jobQueue) Name() string {
	return q.name
}

func (q *jobQueue) Stop() <-chan struct{} {
	return q.impl.Stop()
}

func (q *jobQueue) Push(j IncomingJob) (uint64, error) {
	job, err := q.impl.Push(j)
	if err != nil {
		return 0, err
	}

	q.stats.push(1)

	loggableJob := job.ToLoggable()
	logger.Info(q.name, "push", loggableJob, "New job accepted")

	return loggableJob.ID(), nil
}

func (q *jobQueue) Pop(limit uint) ([]Job, error) {
	results, err := q.impl.Pop(limit)
	if err != nil {
		return nil, err
	}

	q.stats.pop(int64(len(results)))

	for _, j := range results {
		logger.Debug(q.name, "pop", j.ToLoggable(), "A job grabbed")
	}
	return results, nil
}

func (q *jobQueue) Complete(job Job, res *Result) {
	j := &completedJob{job, 0}
	if res.IsFailure() {
		q.stats.fail(1)
		j.failed = 1
	} else {
		q.stats.succeed(1)
	}

	if res.IsPermanentFailure() {
		q.stats.permanentlyFail(1)
	}

	loggable := j.ToLoggable()
	logger.Info(q.name, "complete", loggable, res.Message)

	if res.IsFinished() || !j.canRetry() {
		q.stats.complete(1)
		q.stats.elapsed(logger.Elapsed(loggable))
		if res.IsFailure() {
			if failureLog, ok := q.FailureLog(); ok {
				err := failureLog.Add(job, res)
				if err != nil {
					log.Warn().Msg(err.Error())
				}
			}
		}
		q.impl.Delete(job)
	} else {
		q.impl.Update(job, &nextJob{j})
	}
}

func (q *jobQueue) IsActive() bool {
	return q.impl.IsActive()
}

func (q *jobQueue) Node() (*Node, error) {
	if info, ok := q.impl.(HasNodeInfo); ok {
		return info.Node()
	}
	return nil, nil
}

func (q *jobQueue) Stats() *Stats {
	return q.stats.export()
}

func (q *jobQueue) Inspector() (Inspector, bool) {
	if hasInspector, ok := q.impl.(HasInspector); ok {
		return hasInspector.Inspector(), ok
	}
	return nil, false
}

func (q *jobQueue) FailureLog() (FailureLog, bool) {
	if hasFailureLog, ok := q.impl.(HasFailureLog); ok {
		return hasFailureLog.FailureLog(), ok
	}
	return nil, false
}

// InactiveError is an error returned when Pop() is called on an
// inactive queue.
type InactiveError struct{}

func (e *InactiveError) Error() string {
	return "queue is not active"
}

// ConnectionClosedError is an error returned when Pop() is called but
// connection to a remote store has been lost.
type ConnectionClosedError struct{}

func (e *ConnectionClosedError) Error() string {
	return "connection has been closed"
}
