package inmemory

import (
	"errors"
	"sync"

	"github.com/fireworq/fireworq/model"
	"github.com/fireworq/fireworq/repository"
)

type queueStorage struct {
	sync.RWMutex
	m map[string]model.Queue
}

var qs = &queueStorage{m: make(map[string]model.Queue)}

type queueRepository struct{}

// NewQueueRepository creates a new repository.QueueRepository which
// uses in-memory data store.
func NewQueueRepository() repository.QueueRepository {
	return &queueRepository{}
}

func (r *queueRepository) Add(q *model.Queue) error {
	qs.Lock()
	defer qs.Unlock()

	qs.m[q.Name] = *q

	return nil
}

func (r *queueRepository) FindAll() ([]model.Queue, error) {
	qs.RLock()
	defer qs.RUnlock()

	queues := make([]model.Queue, 0, len(qs.m))
	for _, q := range qs.m {
		queues = append(queues, q)
	}

	return queues, nil
}

func (r *queueRepository) FindByName(name string) (*model.Queue, error) {
	qs.RLock()
	defer qs.RUnlock()

	queue, ok := qs.m[name]
	if !ok {
		return nil, errors.New("Queue not found")
	}
	return &queue, nil
}

func (r *queueRepository) DeleteByName(name string) error {
	qs.Lock()
	defer qs.Unlock()

	delete(qs.m, name)
	return nil
}

func (r *queueRepository) Revision() (uint64, error) {
	return 0, nil
}
