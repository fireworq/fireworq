package jobqueue

import (
	"encoding/json"
	"time"
)

// InspectedJob describes a job in a queue.
type InspectedJob struct {
	ID         uint64          `json:"id"`
	Category   string          `json:"category"`
	URL        string          `json:"url"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
	NextTry    time.Time       `json:"next_try"`
	Timeout    uint            `json:"timeout"`
	FailCount  uint            `json:"fail_count"`
	MaxRetries uint            `json:"max_retries"`
	RetryDelay uint            `json:"retry_delay"`
}

// InspectedJobs describes a (page of) job list in a queue.
type InspectedJobs struct {
	Jobs       []InspectedJob `json:"jobs"`
	NextCursor string         `json:"next_cursor"`
}

// Inspector is an interface to inspect jobs in a queue.
type Inspector interface {
	Delete(jobID uint64) error
	Find(jobID uint64) (*InspectedJob, error)
	FindAllGrabbed(limit uint, cursor string) (*InspectedJobs, error)
	FindAllWaiting(limit uint, cursor string) (*InspectedJobs, error)
	FindAllDeferred(limit uint, cursor string) (*InspectedJobs, error)
}

// HasInspector is an interface describing that it has an Inspector.
//
// This is typically a JobQueue sub-interface.
type HasInspector interface {
	Inspector() Inspector
}

// FailedJob describes a (permanently) failed job that was in a queue.
type FailedJob struct {
	ID        uint64          `json:"id"`
	JobID     uint64          `json:"job_id"`
	Category  string          `json:"category"`
	URL       string          `json:"url"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Result    *Result         `json:"result"`
	FailCount uint            `json:"fail_count"`
	FailedAt  time.Time       `json:"failed_at"`
	CreatedAt time.Time       `json:"created_at"`
}

// FailedJobs describes a (page of) failed job list of a queue.
type FailedJobs struct {
	FailedJobs []FailedJob `json:"failed_jobs"`
	NextCursor string      `json:"next_cursor"`
}

// FailureLog is an interface to inspect failed jobs of a queue.
type FailureLog interface {
	Add(failed Job, result *Result) error
	Delete(failureID uint64) error
	Find(failureID uint64) (*FailedJob, error)
	FindAll(limit uint, cursor string) (*FailedJobs, error)
	FindAllRecentFailures(limit uint, cursor string) (*FailedJobs, error)
}

// HasFailureLog is an interface describing that it has an FailureLog.
//
// This is typically a JobQueue sub-interface.
type HasFailureLog interface {
	FailureLog() FailureLog
}
