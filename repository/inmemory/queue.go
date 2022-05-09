package inmemory

import (
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/fireworq/fireworq/model"
	"github.com/fireworq/fireworq/repository"
)

type queueStorage struct {
	sync.RWMutex
	m        map[string]model.Queue
	revision uint64
}

var qs = &queueStorage{m: make(map[string]model.Queue)}

type queueRepository struct{}

// NewQueueRepository creates a new repository.QueueRepository which
// uses in-memory data store.
func NewQueueRepository() repository.QueueRepository {
	return &queueRepository{}
}

func (r *queueRepository) Add(q *model.Queue) (bool, error) {
	qs.Lock()
	defer qs.Unlock()

	j1, _ := json.Marshal(qs.m[q.Name])
	j2, _ := json.Marshal(q)
	if string(j1) != string(j2) {
		qs.m[q.Name] = *q
		r.updateRevision()
		return true, nil
	}

	return false, nil
}

func (r *queueRepository) FindAll() ([]model.Queue, error) {
	qs.RLock()
	defer qs.RUnlock()

	queues := make([]model.Queue, 0, len(qs.m))
	for _, q := range qs.m {
		queues = append(queues, q)
	}

	sort.Slice(queues, func(i, j int) bool {
		return queues[i].Name < queues[j].Name
	})

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
	r.updateRevision()
	return nil
}

func (r *queueRepository) updateRevision() {
	atomic.AddUint64(&qs.revision, 1)
}

func (r *queueRepository) Revision() (uint64, error) {
	return atomic.LoadUint64(&qs.revision), nil
}
