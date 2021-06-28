package dispatcher

import (
	"context"
	"sync"

	"github.com/fireworq/fireworq/dispatcher/kicker"
	"github.com/fireworq/fireworq/dispatcher/worker"
	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/model"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"golang.org/x/time/rate"
)

const defaultMinBufferSize = 1000
const defaultMinPollingInterval = 100

// Init initializes global parameters of dispatchers by configuration values.
//
// Configuration keys prefixed by "dispatch_" are considered.
func Init() {
	worker.HTTPInit()
}

// Config contains information to create a dispatcher instance.
type Config struct {
	MinBufferSize uint
	Kicker        kicker.Config
	Worker        worker.Config
}

// Start creates and starts a new dispatcher instance with the current
// configuration.
//
// The instance watches a queue specified by q in a way specified by
// m.
func (cfg Config) Start(q JobQueue, m *model.Queue) Dispatcher {
	logger := log.With().Str("package", "dispatcher").Str("queue", q.Name()).Logger()

	bufferSize := cfg.MinBufferSize
	if bufferSize == 0 {
		bufferSize = defaultMinBufferSize
	}
	if m.MaxWorkers > bufferSize {
		bufferSize = m.MaxWorkers
	}

	pollingInterval := m.PollingInterval
	if m.MaxDispatchesPerSecond != 0 {
		pollingInterval = defaultMinPollingInterval
	}
	kc := cfg.Kicker
	if kc == nil {
		kc = &kicker.PollingKicker{Interval: pollingInterval}
	}
	k := kc.NewKicker()

	wc := cfg.Worker
	if wc == nil {
		wc = &worker.HTTPWorker{Logger: &logger}
	}
	w := wc.NewWorker()

	dps := rate.Limit(m.MaxDispatchesPerSecond)
	if dps == 0 {
		dps = rate.Inf
	}
	limiter := rate.NewLimiter(dps, int(m.MaxBurstSize))

	d := &dispatcher{
		jobqueue:  q,
		kicker:    k,
		worker:    w,
		kick:      make(chan struct{}),
		stop:      make(chan struct{}),
		stopped:   make(chan struct{}),
		jobBuffer: make(chan jobqueue.Job, bufferSize),
		sem:       make(chan struct{}, m.MaxWorkers),
		limiter:   limiter,
		logger:    logger,
	}
	go d.loop()
	k.Start(d)

	return d
}

// Dispatcher is an interface of dispatchers for some queue.
type Dispatcher interface {
	Stats() *Stats
	PollingInterval() uint
	MaxWorkers() uint
	MaxDispatchesPerSecond() float64
	MaxBurstSize() int
	Ping()
	Stop() <-chan struct{}
}

// Start creates and starts a new dispatcher instance with the default
// configuration.
func Start(q JobQueue, m *model.Queue) Dispatcher {
	return Config{}.Start(q, m)
}

type dispatcher struct {
	jobqueue  JobQueue
	kicker    kicker.Kicker
	worker    worker.Worker
	kick      chan struct{}
	stop      chan struct{}
	stopped   chan struct{}
	jobBuffer chan jobqueue.Job
	sem       chan struct{}
	limiter   *rate.Limiter
	logger    zerolog.Logger
}

func (d *dispatcher) Kick() {
	d.kick <- struct{}{}
}

func (d *dispatcher) Ping() {
	d.kicker.Ping()
}

func (d *dispatcher) Stats() *Stats {
	runningWorkers := int64(len(d.sem))
	totalWorkers := int64(cap(d.sem))
	return &Stats{
		OutstandingJobs: int64(len(d.jobBuffer)),
		TotalWorkers:    totalWorkers,
		IdleWorkers:     totalWorkers - runningWorkers,
	}
}

func (d *dispatcher) PollingInterval() uint {
	return d.kicker.PollingInterval()
}

func (d *dispatcher) MaxWorkers() uint {
	return uint(cap(d.sem))
}

func (d *dispatcher) MaxDispatchesPerSecond() float64 {
	return float64(d.limiter.Limit())
}

func (d *dispatcher) MaxBurstSize() int {
	return d.limiter.Burst()
}

func (d *dispatcher) Stop() <-chan struct{} {
	stopped := make(chan struct{})

	go func() {
		<-d.kicker.Stop()
		d.stop <- struct{}{}
		<-d.stopped
		stopped <- struct{}{}
	}()

	return stopped
}

func (d *dispatcher) loop() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
Loop:
	for {
		select {
		case <-d.kick:
			d.popJobs()
		case <-d.stop:
			cancel()
			wg.Wait()
			break Loop
		case job := <-d.jobBuffer:
			wg.Add(1)
			d.sem <- struct{}{}
			go func(job jobqueue.Job) {
				defer wg.Done()
				defer func() { <-d.sem }()
				err := d.limiter.Wait(ctx)
				if err == nil {
                    rslt := d.worker.Work(job)
                    d.jobqueue.Complete(job, rslt)
				}
			}(job)
		}
	}
	d.stopped <- struct{}{}
}

func (d *dispatcher) popJobs() {
	if len(d.jobBuffer) < cap(d.jobBuffer) {
		reqn := cap(d.jobBuffer) - len(d.jobBuffer)
		jobs, err := d.jobqueue.Pop(uint(reqn))
		if err != nil {
			switch err.(type) {
			case *jobqueue.InactiveError:
			case *jobqueue.ConnectionClosedError:
			default:
				d.logger.Error().Msgf("Failed to pop jobs: %s", err)
			}
			return
		}
		if len(jobs) > reqn {
			d.logger.Error().Msgf("The number of popped jobs %d is larger than that of requested jobs %d", len(jobs), reqn)
			jobs = jobs[:reqn]
		}
		for _, job := range jobs {
			d.jobBuffer <- job
		}
	}
}

// JobQueue is an interface of a queue which can be watched by
// dispatchers.
type JobQueue interface {
	Pop(limit uint) ([]jobqueue.Job, error)
	Complete(job jobqueue.Job, res *jobqueue.Result)
	Name() string
}

// Stats contains statistics of a dispatcher.
type Stats struct {
	OutstandingJobs int64 `json:"outstanding_jobs"`
	TotalWorkers    int64 `json:"total_workers"`
	IdleWorkers     int64 `json:"idle_workers"`
}
