package mysql

import (
	"database/sql"

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

	return queue, nil
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
