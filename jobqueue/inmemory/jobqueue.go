package inmemory

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/jobqueue/logger"

	"github.com/rs/zerolog/log"
)

type jobQueue struct {
	sync.Mutex
	queue *queue
}

// New creates a jobqueue.Impl which uses in-memory data store.
func New() jobqueue.Impl {
	q := make(queue, 0)
	return &jobQueue{queue: &q}
}

func (q *jobQueue) Start() {
}

func (q *jobQueue) Stop() <-chan struct{} {
	stopped := make(chan struct{}, 1)
	stopped <- struct{}{}
	return stopped
}

func (q *jobQueue) Push(j jobqueue.IncomingJob) (jobqueue.Job, error) {
	q.Lock()
	defer q.Unlock()

	job := newJob(j)
	heap.Push(q.queue, job)
	return job, nil
}

func (q *jobQueue) Pop(limit uint) ([]jobqueue.Job, error) {
	q.Lock()
	defer q.Unlock()

	now := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	popped := make([]jobqueue.Job, 0, limit)
	for i := uint(0); i < limit; i++ {
		if q.queue.Len() <= 0 {
			break
		}
		if (*q.queue)[0].NextTry() > now {
			break
		}

		popped = append(popped, heap.Pop(q.queue).(*job))
	}
	return popped, nil
}

func (q *jobQueue) Delete(job jobqueue.Job) {
	// Do nothing; the job is deleted from the queue on Pop().
}

func (q *jobQueue) Update(completedJob jobqueue.Job, next jobqueue.NextInfo) {
	q.Lock()
	defer q.Unlock()

	j, ok := completedJob.(*job)
	if !ok {
		log.Panic().Msgf("Invalid job structure: %v", completedJob)
		return
	}

	j.nextTry = uint64(time.Now().UnixNano()/int64(time.Millisecond)) + next.NextDelay()
	j.retryCount = next.RetryCount()
	j.failCount = next.FailCount()

	heap.Push(q.queue, j)
}

func (q *jobQueue) IsActive() bool {
	return true
}

type job struct {
	jobqueue.IncomingJob
	id         uint64
	createdAt  uint64
	nextTry    uint64
	retryCount uint
	failCount  uint
}

func newJob(j jobqueue.IncomingJob) *job {
	id := atomic.AddUint64(&lastID, 1)
	createdAt := uint64(time.Now().UnixNano() / int64(time.Millisecond))
	return &job{j, id, createdAt, createdAt + j.NextDelay(), j.RetryCount(), 0}
}

func (j *job) ID() uint64 {
	return j.id
}

func (j *job) CreatedAt() uint64 {
	return j.createdAt
}

func (j *job) Status() string {
	return "claimed"
}

func (j *job) NextTry() uint64 {
	return j.nextTry
}

func (j *job) RetryCount() uint {
	return j.retryCount
}

func (j *job) FailCount() uint {
	return j.failCount
}

func (j *job) ToLoggable() logger.LoggableJob {
	return j
}

type queue []*job

func (q queue) Len() int {
	return len(q)
}

func (q queue) Less(i, j int) bool {
	return q[i].NextTry() < q[j].NextTry()
}

func (q queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *queue) Push(x interface{}) {
	*q = append(*q, x.(*job))
}

func (q *queue) Pop() interface{} {
	old := *q
	n := len(old)
	x := old[n-1]
	*q = old[0 : n-1]
	return x
}

var lastID uint64
