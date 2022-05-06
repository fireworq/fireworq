package inmemory

import (
	"sync"
	"sync/atomic"

	"github.com/fireworq/fireworq/model"
	"github.com/fireworq/fireworq/repository"
)

type routingStorage struct {
	sync.RWMutex
	m        map[string]string
	revision uint64
}

var rs = &routingStorage{m: make(map[string]string)}

type routingRepository struct{}

// NewRoutingRepository creates a new repository.RoutingRepository
// which uses in-memory data store.
func NewRoutingRepository() repository.RoutingRepository {
	return &routingRepository{}
}

func (r *routingRepository) Add(jobCategory string, queueName string) (bool, error) {
	rs.Lock()
	defer rs.Unlock()

	if rs.m[jobCategory] != queueName {
		rs.m[jobCategory] = queueName
		return true, nil
	}
	return false, nil
}

func (r *routingRepository) FindAll() ([]model.Routing, error) {
	rs.RLock()
	defer rs.RUnlock()

	routings := make([]model.Routing, 0, len(rs.m))
	for category, queue := range rs.m {
		routings = append(routings, model.Routing{
			QueueName:   queue,
			JobCategory: category,
		})
	}

	return routings, nil
}

func (r *routingRepository) FindQueueNameByJobCategory(category string) string {
	rs.RLock()
	defer rs.RUnlock()

	return rs.m[category]
}

func (r *routingRepository) DeleteByJobCategory(category string) error {
	rs.RLock()
	defer rs.RUnlock()

	delete(rs.m, category)
	return nil
}

func (r *routingRepository) updateRevision() {
	atomic.AddUint64(&rs.revision, 1)
}

func (r *routingRepository) Revision() (uint64, error) {
	return atomic.LoadUint64(&rs.revision), nil
}

func (r *routingRepository) Reload() error {
	return nil
}
