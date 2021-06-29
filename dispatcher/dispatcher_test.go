package dispatcher

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fireworq/fireworq/dispatcher/kicker"
	"github.com/fireworq/fireworq/dispatcher/worker"
	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/jobqueue/logger"
	"github.com/fireworq/fireworq/model"

	"golang.org/x/time/rate"
)

func TestMain(m *testing.M) {
	Init()
	os.Exit(m.Run())
}

func TestStart(t *testing.T) {
	{
		pollingInterval := uint(500)
		maxWorkers := uint(5)
		cfg := &Config{}
		d := cfg.Start(&dummyJobQueue{}, &model.Queue{
			PollingInterval: pollingInterval,
			MaxWorkers:      maxWorkers,
		})
		defer func() { <-d.Stop() }()

		if d.PollingInterval() != pollingInterval {
			t.Errorf("Wrong polling interval: %d", d.PollingInterval())
		}

		if d.MaxWorkers() != maxWorkers {
			t.Errorf("Wrong max workers: %d", d.MaxWorkers())
		}

		if d.MaxDispatchesPerSecond() != float64(rate.Inf) {
			t.Errorf("Wrong max dispatches per second: %f", d.MaxDispatchesPerSecond())
		}

		if d.MaxBurstSize() != 0 {
			t.Errorf("Wrong max burst size: %d", d.MaxBurstSize())
		}
	}

	{
		pollingInterval := uint(500)
		maxWorkers := uint(5)
		d := Start(&dummyJobQueue{}, &model.Queue{
			PollingInterval: pollingInterval,
			MaxWorkers:      maxWorkers,
		})
		defer func() { <-d.Stop() }()

		if d.PollingInterval() != pollingInterval {
			t.Errorf("Wrong polling interval: %d", d.PollingInterval())
		}

		if d.MaxWorkers() != maxWorkers {
			t.Errorf("Wrong max workers: %d", d.MaxWorkers())
		}

		if d.MaxDispatchesPerSecond() != float64(rate.Inf) {
			t.Errorf("Wrong max dispatches per second: %f", d.MaxDispatchesPerSecond())
		}

		if d.MaxBurstSize() != 0 {
			t.Errorf("Wrong max burst size: %d", d.MaxBurstSize())
		}
	}

	{
		pollingInterval := uint(100)
		maxWorkers := uint(5)
		maxDispatchesPerSecond := float64(2.0)
		maxBurstSize := uint(3)
		d := Start(&dummyJobQueue{}, &model.Queue{
			PollingInterval:        pollingInterval,
			MaxWorkers:             maxWorkers,
			MaxDispatchesPerSecond: maxDispatchesPerSecond,
			MaxBurstSize:           maxBurstSize,
		})
		defer func() { <-d.Stop() }()

		if d.PollingInterval() != pollingInterval {
			t.Errorf("Wrong polling interval: %d", d.PollingInterval())
		}

		if d.MaxWorkers() != maxWorkers {
			t.Errorf("Wrong max workers: %d", d.MaxWorkers())
		}

		if d.MaxDispatchesPerSecond() != maxDispatchesPerSecond {
			t.Errorf("Wrong max dispatches per second: %f", d.MaxDispatchesPerSecond())
		}

		if d.MaxBurstSize() != int(maxBurstSize) {
			t.Errorf("Wrong max burst size: %d", d.MaxBurstSize())
		}
	}
}

func TestStop(t *testing.T) {
	kicker := &dummyKicker{}

	cfg := Config{Kicker: &dummyKickerConfig{instance: kicker}}
	d := cfg.Start(&dummyJobQueue{}, &model.Queue{})

	<-d.Stop()
	if kicker.stopped != 1 {
		t.Error("Kicker must be stopped")
	}
}

func TestKick(t *testing.T) {
	kicker := &dummyKicker{}

	jobs := make([]jobqueue.Job, 0)
	for i := 0; i < 15; i++ {
		jobs = append(jobs, &job{fmt.Sprintf("%d", i)})
	}
	jq := &dummyJobQueue{jobs: jobs}

	cfg := Config{
		MinBufferSize: 10,
		Kicker:        &dummyKickerConfig{instance: kicker},
		Worker:        &dummyWorker{},
	}
	d := cfg.Start(jq, &model.Queue{MaxWorkers: 1}).(*dispatcher)
	defer func() { <-d.Stop() }()

	if len(jq.completed) != 0 {
		t.Error("Queue must not be popped without kicking")
	}

	d.Kick()
	time.Sleep(200 * time.Millisecond)

	func() {
		jq.Lock()
		defer jq.Unlock()

		if len(jq.completed) != 10 {
			t.Error("Queue must be popped on kicking")
		}
	}()

	d.Kick()
	time.Sleep(200 * time.Millisecond)

	func() {
		jq.Lock()
		defer jq.Unlock()

		if len(jq.completed) != 15 {
			t.Error("Queue must be popped on kicking")
		}

		for i, r := range jq.completed {
			if r.Message != fmt.Sprintf("%d", i) {
				t.Error("Jobs must be processed FIFO")
			}
		}

		if len(jq.jobs) != 0 {
			t.Error("jobs must be popped")
		}
	}()
}

func TestWorkConcurrently(t *testing.T) {
	kicker := &dummyKicker{}

	jobs := make([]jobqueue.Job, 0)
	for i := 0; i < 15; i++ {
		jobs = append(jobs, &job{fmt.Sprintf("%d", i)})
	}
	jq := &dummyJobQueue{jobs: jobs}

	cfg := Config{
		MinBufferSize: 1,
		Kicker:        &dummyKickerConfig{instance: kicker},
		Worker:        &dummyWorker{},
	}
	d := cfg.Start(jq, &model.Queue{MaxWorkers: 20}).(*dispatcher)
	defer func() { <-d.Stop() }()

	if len(jq.completed) != 0 {
		t.Error("Queue must not be popped without kicking")
	}

	d.Kick()
	time.Sleep(300 * time.Millisecond)

	jq.Lock()
	defer jq.Unlock()

	if len(jq.completed) != 15 {
		t.Error("Queue must be popped on kicking")
	}
}

func TestThrottling(t *testing.T) {
	kicker := &dummyKicker{}

	jobs := make([]jobqueue.Job, 0)
	for i := 0; i < 15; i++ {
		jobs = append(jobs, &job{fmt.Sprintf("%d", i)})
	}
	jq := &dummyJobQueue{jobs: jobs}

	cfg := Config{
		MinBufferSize: 10,
		Kicker:        &dummyKickerConfig{instance: kicker},
		Worker:        &dummyWorker{},
	}
	d := cfg.Start(jq, &model.Queue{MaxWorkers: 20, MaxDispatchesPerSecond: 0.000001, MaxBurstSize: 1}).(*dispatcher)

	if len(jq.completed) != 0 {
		t.Error("Queue must not be popped without kicking")
	}

	d.Kick()
	time.Sleep(300 * time.Millisecond)

	jq.Lock()
	if len(jq.completed) != 1 || len(jq.jobs) != 0 {
		t.Error("Jobs must be throttled")
	}
	jq.Unlock()

	<-d.Stop()
	if kicker.stopped != 1 {
		t.Error("Kicker must be stopped")
	}

	jq.Lock()
	defer jq.Unlock()

	if len(jq.completed) != 1 {
		t.Error("Canceled jobs must not be completed")
	}
}

func TestSkipPopping(t *testing.T) {
	kicker := &dummyKicker{}

	cfg := Config{
		Kicker: &dummyKickerConfig{instance: kicker},
		Worker: &dummyWorker{},
	}

	func() {
		jq := &errorJobQueue{err: &jobqueue.InactiveError{}}
		d := cfg.Start(jq, &model.Queue{}).(*dispatcher)
		defer func() { <-d.Stop() }()

		if atomic.LoadInt64(&jq.completed) != 0 {
			t.Error("Queue must not be popped without kicking")
		}

		d.Kick()
		time.Sleep(200 * time.Millisecond)

		if atomic.LoadInt64(&jq.completed) != 0 {
			t.Error("There must be no job to complete")
		}
	}()

	func() {
		jq := &errorJobQueue{err: &jobqueue.ConnectionClosedError{}}
		d := cfg.Start(jq, &model.Queue{}).(*dispatcher)
		defer func() { <-d.Stop() }()

		if atomic.LoadInt64(&jq.completed) != 0 {
			t.Error("Queue must not be popped without kicking")
		}

		d.Kick()
		time.Sleep(200 * time.Millisecond)

		if atomic.LoadInt64(&jq.completed) != 0 {
			t.Error("There must be no job to complete")
		}
	}()

	func() {
		jq := &errorJobQueue{err: errors.New("Some other error")}
		d := cfg.Start(jq, &model.Queue{}).(*dispatcher)
		defer func() { <-d.Stop() }()

		if atomic.LoadInt64(&jq.completed) != 0 {
			t.Error("Queue must not be popped without kicking")
		}

		d.Kick()
		time.Sleep(200 * time.Millisecond)

		if atomic.LoadInt64(&jq.completed) != 0 {
			t.Error("There must be no job to complete")
		}
	}()
}

func TestBrokenJobQueue(t *testing.T) {
	kicker := &dummyKicker{}

	jobs := make([]jobqueue.Job, 0)
	for i := 0; i < 15; i++ {
		jobs = append(jobs, &job{fmt.Sprintf("%d", i)})
	}
	jq := &brokenJobQueue{dummyJobQueue{jobs: jobs}}

	cfg := Config{
		MinBufferSize: 10,
		Kicker:        &dummyKickerConfig{instance: kicker},
		Worker:        &dummyWorker{},
	}
	d := cfg.Start(jq, &model.Queue{MaxWorkers: 1}).(*dispatcher)
	defer func() { <-d.Stop() }()

	if len(jq.completed) != 0 {
		t.Error("Queue must not be popped without kicking")
	}

	d.Kick()
	time.Sleep(200 * time.Millisecond)

	func() {
		jq.Lock()
		defer jq.Unlock()

		if len(jq.completed) != 10 {
			t.Error("Queue must be popped on kicking")
		}
	}()
}

func TestPing(t *testing.T) {
	kicker := &dummyKicker{}

	cfg := Config{Kicker: &dummyKickerConfig{instance: kicker}}
	d := cfg.Start(&dummyJobQueue{}, &model.Queue{})
	defer func() { <-d.Stop() }()

	d.Ping()
	if kicker.pinged != 1 {
		t.Error("Kicker must be pinged")
	}
}

func TestStats(t *testing.T) {
	worker := &dummyBlockingWorker{make(chan struct{}, 1)}

	kicker := &dummyKicker{}

	totalJobs := 10
	minBufferSize := uint(8)
	maxWorkers := uint(3)

	jobs := make([]jobqueue.Job, 0)
	for i := 0; i < totalJobs; i++ {
		jobs = append(jobs, &job{fmt.Sprintf("%d", i)})
	}
	jq := &dummyJobQueue{jobs: jobs}

	cfg := Config{
		MinBufferSize: minBufferSize,
		Kicker:        &dummyKickerConfig{instance: kicker},
		Worker:        worker,
	}
	d := cfg.Start(jq, &model.Queue{MaxWorkers: maxWorkers}).(*dispatcher)
	defer func() { <-d.Stop() }()

	if len(jq.completed) != 0 {
		t.Error("Queue must not be popped without kicking")
	}

	if d.Stats().IdleWorkers != int64(maxWorkers) {
		t.Error("Workers should be idle before handling jobs")
	}

	d.Kick()
	time.Sleep(300 * time.Millisecond)

	stats := d.Stats()
	if stats.OutstandingJobs+1 < int64(minBufferSize-maxWorkers) {
		t.Error("Popped jobs should be stored in a buffer")
	}
	if stats.IdleWorkers != 0 {
		t.Error("Popped jobs should be handled by workers")
	}

	for i := maxWorkers; i < uint(totalJobs); i += maxWorkers {
		go d.Kick()
	}
	for i := 0; i < totalJobs; i++ {
		worker.Process()
	}
}

type dummyJobQueue struct {
	sync.Mutex
	jobs      []jobqueue.Job
	completed []jobqueue.Result
}

func (jq *dummyJobQueue) Pop(limit uint) ([]jobqueue.Job, error) {
	jq.Lock()
	defer jq.Unlock()

	if len(jq.jobs) > int(limit) {
		grabbed, rest := jq.jobs[0:limit], jq.jobs[limit:]
		jq.jobs = rest
		return grabbed, nil
	}

	grabbed := jq.jobs
	jq.jobs = jq.jobs[:0]
	return grabbed, nil
}

func (jq *dummyJobQueue) Complete(job jobqueue.Job, res *jobqueue.Result) {
	jq.Lock()
	defer jq.Unlock()

	jq.completed = append(jq.completed, *res)
}

func (jq *dummyJobQueue) Name() string { return "dummy" }

type errorJobQueue struct {
	err       error
	completed int64
}

func (jq *errorJobQueue) Pop(limit uint) ([]jobqueue.Job, error) {
	return nil, jq.err
}

func (jq *errorJobQueue) Name() string { return "error" }

func (jq *errorJobQueue) Complete(job jobqueue.Job, res *jobqueue.Result) {
	atomic.AddInt64(&jq.completed, 1)
}

type brokenJobQueue struct {
	dummyJobQueue
}

func (jq *brokenJobQueue) Pop(limit uint) ([]jobqueue.Job, error) {
	return jq.dummyJobQueue.Pop(limit + 1)
}

type dummyKicker struct {
	pinged  int
	stopped int
}

func (k *dummyKicker) Start(kickable kicker.Kickable) {}

func (k *dummyKicker) Stop() <-chan struct{} {
	stopped := make(chan struct{}, 1)
	k.stopped++
	stopped <- struct{}{}
	return stopped
}

func (k *dummyKicker) Ping() {
	k.pinged++
}

func (k *dummyKicker) PollingInterval() uint { return 0 }

type dummyKickerConfig struct {
	instance kicker.Kicker
}

func (k *dummyKickerConfig) NewKicker() kicker.Kicker {
	return k.instance
}

type dummyWorker struct{}

func (w *dummyWorker) Work(job jobqueue.Job) *jobqueue.Result {
	return &jobqueue.Result{
		Status:  jobqueue.ResultStatusSuccess,
		Message: job.Payload(),
	}
}

func (w *dummyWorker) NewWorker() worker.Worker { return w }

type dummyBlockingWorker struct {
	ch chan struct{}
}

func (w *dummyBlockingWorker) NewWorker() worker.Worker { return w }

func (w *dummyBlockingWorker) Process() {
	w.ch <- struct{}{}
}

func (w *dummyBlockingWorker) Work(job jobqueue.Job) *jobqueue.Result {
	<-w.ch
	return &jobqueue.Result{
		Status:  jobqueue.ResultStatusSuccess,
		Message: job.Payload(),
	}
}

type job struct {
	payload string
}

func (j *job) URL() string                    { return "" }
func (j *job) Payload() string                { return j.payload }
func (j *job) RetryCount() uint               { return 0 }
func (j *job) RetryDelay() uint               { return 0 }
func (j *job) FailCount() uint                { return 0 }
func (j *job) Timeout() uint                  { return 0 }
func (j *job) ToLoggable() logger.LoggableJob { return nil }
