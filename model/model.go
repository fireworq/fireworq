package model

// Queue describes a queue.
type Queue struct {
	Name            string `json:"name"`
	PollingInterval uint   `json:"polling_interval"`
	MaxWorkers      uint   `json:"max_workers"`
}

// Routing describes a routing.
type Routing struct {
	QueueName   string `json:"queue_name"`
	JobCategory string `json:"job_category"`
}
