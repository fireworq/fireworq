package service

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/fireworq/fireworq/config"
	jobqueue "github.com/fireworq/fireworq/jobqueue/factory"
	"github.com/fireworq/fireworq/model"
	"github.com/fireworq/fireworq/repository"

	"github.com/rs/zerolog/log"
)

// PushResult is information of pushed job except those in
// jobqueue.IncomingJob itself.
type PushResult struct {
	ID        uint64
	QueueName string
}

// Service is an application use case service that manages running
// queues.
type Service struct {
	defaultQueueName string
	queue            repository.QueueRepository
	routing          repository.RoutingRepository
	runningQueues    map[string]RunningQueue
	mu               sync.Mutex
	muJob            sync.RWMutex
	queueW           *configWatcher
	routingW         *configWatcher
}

// NewService creates a new Service instance.
func NewService(repos *repository.Repositories) *Service {
	s := &Service{
		defaultQueueName: config.Get("queue_default"),
		queue:            repos.Queue,
		routing:          repos.Routing,
		runningQueues:    make(map[string]RunningQueue),
	}
	s.queueW = newConfigWatcher(
		s.queue.Revision,
		s.reloadQueues,
	)
	s.routingW = newConfigWatcher(
		s.routing.Revision,
		s.reloadRoutings,
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.muJob.Lock()
	defer s.muJob.Unlock()

	s.startup()
	s.queueW.start(configRefreshInterval())
	s.routingW.start(configRefreshInterval())

	return s
}

// Stop stops all the running queues.
//
// This method should not be called more than once in the whole
// application.
func (s *Service) Stop() <-chan struct{} {
	stopped := make(chan struct{})
	go func() {
		<-s.queueW.stop()
		<-s.routingW.stop()

		s.mu.Lock()
		defer s.mu.Unlock()

		s.deactivateQueues()

		s.muJob.Lock()
		defer s.muJob.Unlock()

		s.destroyQueues()

		stopped <- struct{}{}
	}()
	return stopped
}

// GetJobQueue returns a RunningQueue of name qn.  The second return
// value is false ff no queue is found.
//
// This method is goroutine safe.
func (s *Service) GetJobQueue(qn string) (RunningQueue, bool) {
	s.muJob.RLock()
	defer s.muJob.RUnlock()

	q, ok := s.getJobQueue(qn)
	return q, ok
}

func (s *Service) getJobQueue(qn string) (RunningQueue, bool) {
	jobQueue, ok := s.runningQueues[qn]
	return jobQueue, ok
}

// DeleteJobQueue stops a running queue of name qn and removes it from
// the queue definition list.
//
// This method is goroutine safe.
func (s *Service) DeleteJobQueue(qn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.muJob.Lock()
	defer s.muJob.Unlock()

	if err := s.queue.DeleteByName(qn); err != nil {
		return err
	}

	if jq, ok := s.runningQueues[qn]; ok {
		<-jq.Deactivate()
		<-jq.Stop()
		delete(s.runningQueues, qn)
	}

	return nil
}

// AddJobQueue defines a new queue and starts it.
//
// This method is goroutine safe.
func (s *Service) AddJobQueue(q *model.Queue) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.muJob.Lock()
	defer s.muJob.Unlock()

	return s.addJobQueue(q)
}

func (s *Service) addJobQueue(q *model.Queue) error {
	if q.PollingInterval == 0 {
		q.PollingInterval = defaultPollingInterval()
	}
	if q.MaxWorkers == 0 {
		q.MaxWorkers = defaultMaxWorkers()
	}

	err := s.queue.Add(q)
	if err != nil {
		return err
	}
	s.putJobQueue(q)

	return nil
}

// Push pushes a job to a queue.  The target queue is determined by
// the category of the job and defined routings.
func (s *Service) Push(job jobqueue.IncomingJob) (*PushResult, error) {
	qn := s.routing.FindQueueNameByJobCategory(job.Category())
	if qn == "" {
		qn = s.defaultQueueName
	}
	if qn == "" {
		s.routing.Reload()
		qn = s.routing.FindQueueNameByJobCategory(job.Category())
	}
	if qn == "" {
		return nil, fmt.Errorf("No routing of job category '%s' exists", job.Category())
	}

	ok, id, err := func() (bool, uint64, error) {
		s.muJob.RLock()
		defer s.muJob.RUnlock()

		jq, ok := s.getJobQueue(qn)
		if !ok {
			return ok, 0, nil
		}

		id, err := jq.Push(job)
		return ok, id, err
	}()
	if err != nil {
		return nil, err
	}

	if !ok {
		// This happens when the queue definition is not in the cache
		// but in the data store, which means it has been defined
		// through another node.

		s.mu.Lock()
		defer s.mu.Unlock()

		s.muJob.Lock()
		defer s.muJob.Unlock()

		q, err := s.queue.FindByName(qn)
		if err != nil {
			return nil, fmt.Errorf("Undefined queue: %s", qn)
		}
		jq := s.putJobQueue(q)

		id, err = jq.Push(job)
		if err != nil {
			return nil, err
		}
	}

	return &PushResult{ID: id, QueueName: qn}, nil
}

func (s *Service) startup() {
	qs, err := s.queue.FindAll()
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	for _, q := range qs {
		s.putJobQueue(&q)
	}

	queueName := config.Get("queue_default")
	if len(queueName) > 0 {
		err := s.initDefaultQueue(queueName)
		if err != nil {
			log.Panic().Msgf("Cannot create default job queue: %s", queueName)
		}
	}

	log.Info().Msgf("Started %d queue dispatchers", len(s.runningQueues))
}

func (s *Service) reloadQueues() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Info().Msg("Deactivating queues...")
	s.deactivateQueues()

	s.muJob.Lock()
	defer s.muJob.Unlock()

	log.Info().Msg("Reloading queue definitions...")
	s.destroyQueues()
	s.startup()
}

func (s *Service) reloadRoutings() {
	log.Info().Msg("Reloading routings...")
	s.routing.Reload()
}

func (s *Service) initDefaultQueue(queueName string) error {
	_, ok := s.getJobQueue(queueName)
	if !ok {
		queue := &model.Queue{Name: queueName}
		err := s.addJobQueue(queue)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) putJobQueue(q *model.Queue) RunningQueue {
	if jq, ok := s.runningQueues[q.Name]; ok {
		<-jq.Deactivate()
		<-jq.Stop()
		delete(s.runningQueues, q.Name)
	}

	jq := startJobQueue(q)
	s.runningQueues[q.Name] = jq
	return jq
}

func (s *Service) deactivateQueues() {
	n := len(s.runningQueues)
	ch := make(chan struct{}, n)
	for _, q := range s.runningQueues {
		go func(q RunningQueue) {
			<-q.Deactivate()
			ch <- struct{}{}
		}(q)
	}
	for n > 0 {
		<-ch
		n--
	}
}

func (s *Service) destroyQueues() {
	n := len(s.runningQueues)
	ch := make(chan struct{}, n)
	for _, q := range s.runningQueues {
		go func(q RunningQueue) {
			<-q.Stop()
			ch <- struct{}{}
		}(q)
	}
	for n > 0 {
		<-ch
		n--
	}
	s.runningQueues = make(map[string]RunningQueue)
}

func defaultPollingInterval() uint {
	str := config.Get("queue_default_polling_interval")
	pollingInterval, err := strconv.Atoi(str)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	return uint(pollingInterval)
}

func defaultMaxWorkers() uint {
	str := config.Get("queue_default_max_workers")
	pollingInterval, err := strconv.Atoi(str)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	return uint(pollingInterval)
}

func configRefreshInterval() uint {
	str := config.Get("config_refresh_interval")
	interval, err := strconv.Atoi(str)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	return uint(interval)
}
