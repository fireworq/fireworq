package mysql

import (
	"database/sql"
	"strings"

	"github.com/fireworq/fireworq/model"
	"github.com/fireworq/fireworq/repository"
)

type queueRepository struct {
	db *sql.DB
}

// NewQueueRepository creates a repository.QueueRepository which uses
// MySQL as a data store.
func NewQueueRepository(db *sql.DB) repository.QueueRepository {
	return &queueRepository{db: db}
}

func (r *queueRepository) Add(q *model.Queue) error {
	sql := `
		INSERT INTO queue (name, polling_interval, max_workers)
		VALUES ( ?, ?, ? )
		ON DUPLICATE KEY UPDATE
			polling_interval = VALUES(polling_interval),
			max_workers = VALUES(max_workers)
	`
	_, err := r.db.Exec(sql, q.Name, q.PollingInterval, q.MaxWorkers)
	if err != nil {
		return err
	}

	sql = `
		INSERT INTO queue_throttle (name, max_dispatches_per_second, max_burst_size)
		VALUES ( ?, ?, ? )
		ON DUPLICATE KEY UPDATE
			max_dispatches_per_second = VALUES(max_dispatches_per_second),
			max_burst_size = VALUES(max_burst_size)
	`
	_, err = r.db.Exec(sql, q.Name, q.MaxDispatchesPerSecond, q.MaxBurstSize)
	if err != nil {
		return err
	}

	return r.updateRevision()
}

func (r *queueRepository) FindAll() ([]model.Queue, error) {
	sql := `
		SELECT name, polling_interval, max_workers
		FROM queue
		ORDER BY name ASC
	`
	rows, err := r.db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]model.Queue, 0)
	for rows.Next() {
		var q model.Queue
		if err := rows.Scan(&(q.Name), &(q.PollingInterval), &(q.MaxWorkers)); err != nil {
			return nil, err
		}
		results = append(results, q)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	names := make([]string, len(results))
	for i, q := range results {
		names[i] = q.Name
	}
	throttles, err := r.findQueueThrottles(names)
	if err != nil {
		return nil, err
	}
	for i, q := range results {
		if throttle, ok := throttles[q.Name]; ok {
			results[i].MaxDispatchesPerSecond = throttle.maxDispatchesPerSecond
			results[i].MaxBurstSize = throttle.maxBurstSize
		}
	}

	return results, nil
}

func (r *queueRepository) FindByName(name string) (*model.Queue, error) {
	sql := `
		SELECT name, polling_interval, max_workers FROM queue
		WHERE name = ?
	`

	queue := &model.Queue{}
	err := r.db.QueryRow(sql, name).Scan(
		&(queue.Name),
		&(queue.PollingInterval),
		&(queue.MaxWorkers),
	)
	if err != nil {
		return nil, err
	}

	throttles, err := r.findQueueThrottles([]string{queue.Name})
	if err != nil {
		return nil, err
	}
	if throttle, ok := throttles[queue.Name]; ok {
		queue.MaxDispatchesPerSecond = throttle.maxDispatchesPerSecond
		queue.MaxBurstSize = throttle.maxBurstSize
	}

	return queue, nil
}

type queueThrottle struct {
	maxDispatchesPerSecond float64
	maxBurstSize           uint
}

func (r *queueRepository) findQueueThrottles(names []string) (map[string]queueThrottle, error) {
	if len(names) == 0 {
		return nil, nil
	}

	sql := `
		SELECT name, max_dispatches_per_second, max_burst_size
		FROM queue_throttle
		WHERE name IN (` + strings.Repeat("?,", len(names)-1) + `?)
	`

	args := make([]interface{}, len(names))
	for i, name := range names {
		args[i] = name
	}

	rows, err := r.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		name                   string
		maxDispatchesPerSecond float64
		maxBurstSize           uint
		throttleByName         = make(map[string]queueThrottle, len(names))
	)
	for rows.Next() {
		if err := rows.Scan(&name, &maxDispatchesPerSecond, &maxBurstSize); err != nil {
			return nil, err
		}
		throttleByName[name] = queueThrottle{maxDispatchesPerSecond, maxBurstSize}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return throttleByName, nil
}

func (r *queueRepository) DeleteByName(name string) error {
	sql := `
		DELETE FROM queue
		WHERE name = ?
	`
	_, err := r.db.Exec(sql, name)
	if err != nil {
		return err
	}

	sql = `
		DELETE FROM queue_throttle
		WHERE name = ?
	`
	_, err = r.db.Exec(sql, name)
	if err != nil {
		return err
	}

	return r.updateRevision()
}

func (r *queueRepository) Revision() (uint64, error) {
	var revision uint64
	if err := r.db.QueryRow(`
		SELECT revision FROM config_revision
		WHERE name = 'queue_definition'
	`).Scan(&revision); err != nil {
		return 0, err
	}
	return revision, nil
}

func (r *queueRepository) updateRevision() error {
	_, err := r.db.Exec(`
		INSERT INTO config_revision (name, revision)
		VALUES ('queue_definition', 1)
		ON DUPLICATE KEY UPDATE
			revision = revision + 1
	`)
	return err
}
