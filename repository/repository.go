package repository

import "github.com/fireworq/fireworq/model"

// QueueRepository is an interface of a queue repository.
type QueueRepository interface {
	Add(q *model.Queue) error
	FindAll() ([]model.Queue, error)
	FindByName(name string) (*model.Queue, error)
	DeleteByName(name string) error
	Revision() (uint64, error)
}

// RoutingRepository is an interface of a routing repository.
type RoutingRepository interface {
	Add(jobCategory string, queueName string) error
	FindAll() ([]model.Routing, error)
	FindQueueNameByJobCategory(category string) string
	DeleteByJobCategory(category string) error
	Revision() (uint64, error)
	Reload() error
}

// Repositories contains a queue repository and a routing repository.
type Repositories struct {
	Queue   QueueRepository
	Routing RoutingRepository
}
