package model

// Queue describes a queue.
type Queue struct {
	Name                   string  `json:"name"`
	PollingInterval        uint    `json:"polling_interval,omitempty"`
	MaxWorkers             uint    `json:"max_workers"`
	MaxDispatchesPerSecond float64 `json:"max_dispatches_per_second,omitempty"`
	MaxBurstSize           uint    `json:"max_burst_size,omitempty"`
}

// Routing describes a routing.
type Routing struct {
	QueueName   string `json:"queue_name"`
	JobCategory string `json:"job_category"`
}
