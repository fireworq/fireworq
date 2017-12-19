package inmemory

import (
	"sync"

	"github.com/fireworq/fireworq/model"
	"github.com/fireworq/fireworq/repository"
)

type routingStorage struct {
	sync.RWMutex
	m map[string]string
}

var rs = &routingStorage{m: make(map[string]string)}

type routingRepository struct{}

// NewRoutingRepository creates a new repository.RoutingRepository
// which uses in-memory data store.
func NewRoutingRepository() repository.RoutingRepository {
	return &routingRepository{}
}

func (r *routingRepository) Add(jobCategory string, queueName string) error {
	rs.Lock()
	defer rs.Unlock()

	rs.m[jobCategory] = queueName
	return nil
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

func (r *routingRepository) Revision() (uint64, error) {
	return 0, nil
}

func (r *routingRepository) Reload() error {
	return nil
}
