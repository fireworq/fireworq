package mysql

import (
	"context"
	"database/sql"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql" // initialize the driver
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/fireworq/fireworq/config"
	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/model"
)

// Dsn returns the data source name of the storage specified in the
// configuration.
func Dsn() string {
	dsn := config.Get("queue_mysql_dsn")
	if dsn != "" {
		return dsn
	}
	return config.Get("mysql_dsn")
}

type jobQueue struct {
	name    string
	dsn     string
	sql     *sqls
	db      *sql.DB
	dbPop   *sql.DB
	mu      sync.RWMutex
	stopped uint32
	logger  zerolog.Logger
}

// New creates a jobqueue.Impl which uses MySQL as a data store.
func New(definition *model.Queue, dsn string) jobqueue.Impl {
	return newJobQueue(definition, dsn)
}

func newJobQueue(definition *model.Queue, dsn string) *jobQueue {
	tableName := newTableName(definition)
	return &jobQueue{
		name:   definition.Name,
		dsn:    dsn,
		sql:    tableName.makeQueries(),
		logger: log.With().Str("queue", definition.Name).Logger(),
	}
}

func (q *jobQueue) Start() {
	log := q.logger.With().Str("method", "Start").Logger()

	db, err := sql.Open("mysql", q.dsn)
	if err != nil {
		log.Panic().Msgf("Cannot open DB: %s", err)
	}
	func() {
		var timeout int
		if db.QueryRow("SELECT @@SESSION.wait_timeout").Scan(&timeout) != nil {
			return
		}

		t := timeout - 1
		if t < 1 {
			t = 1
		}
		log.Debug().Msgf("wait_timeout: %d", timeout)
		db.SetConnMaxLifetime(time.Duration(t) * time.Second)
	}()
	q.db = db

	_, err = q.db.Exec(q.sql.createJobqueue)
	if err != nil {
		log.Panic().Msgf("Failed to create queue table: %s", err)
	}

	_, err = q.db.Exec(q.sql.createFailure)
	if err != nil {
		log.Panic().Msgf("Failed to create queue failure log table: %s", err)
	}

	q.connect()
}

func (q *jobQueue) Stop() <-chan struct{} {
	atomic.StoreUint32(&q.stopped, 1)

	stopped := make(chan struct{})
	go func() {
		q.disconnect()
		q.db.Close()
		stopped <- struct{}{}
	}()
	return stopped
}

func (q *jobQueue) IsActive() bool {
	return true
}

func (q *jobQueue) Push(j jobqueue.IncomingJob) (jobqueue.Job, error) {
	log := q.logger.With().Str("method", "Push").Logger()

	job := &incomingJob{j, 0}

	r, err := q.db.Exec(
		q.sql.insertJob,
		job.NextDelay(),
		job.RetryCount(),
		job.RetryDelay(),
		job.FailCount(),
		job.Category(),
		job.URL(),
		job.Payload(),
		job.Timeout(),
	)
	if err != nil {
		log.Debug().Msgf("Failed to insert a job: %s", err)
		return nil, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		log.Debug().Msgf("Cannot get the last insert ID of the new job: %s", err)
		return nil, err
	}
	job.id = uint64(id)

	return job, nil
}

func (q *jobQueue) Pop(limit uint) ([]jobqueue.Job, error) {
	log := q.logger.With().Str("method", "Pop").Logger()

	if !q.IsActive() {
		return nil, &jobqueue.InactiveError{}
	}

	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.dbPop == nil {
		return nil, &jobqueue.ConnectionClosedError{}
	}

	placeholders := make([]string, 0, limit)
	ids := make([]interface{}, 0, limit)

	// 1. Pre-SELECT jobs to grab.  We should not `SELECT ~ FOR UPDATE`
	// here because it blocks `Push()` due to a gap lock.
	if err := func() error {
		rows, err := q.dbPop.Query(q.sql.grab + strconv.FormatUint(uint64(limit), 10))
		if err != nil {
			log.Debug().Msgf("Failed to preselect jobs: %s", err)
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id uint64
			if err := rows.Scan(&id); err != nil {
				log.Debug().Msgf("Failed to scan preselected job IDs: %s", err)
				return err
			}
			placeholders = append(placeholders, "?")
			ids = append(ids, id)
		}
		if err := rows.Err(); err != nil {
			log.Debug().Msgf("Failed to read preselected job IDs: %s", err)
			return err
		}
		return nil
	}(); err != nil {
		return nil, err
	}
	if len(ids) <= 0 { // no job to grab
		return []jobqueue.Job{}, nil
	}

	results := make([]jobqueue.Job, 0, limit)
	ctx := context.Background()
	tx, err := q.dbPop.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted, // to avoid gap locks
		ReadOnly:  false,
	})
	if err != nil {
		return nil, err
	}

	// 2. SELECT jobs to grab FOR UPDATEing their status.
	if err := func() error {
		rows, err := tx.Query(
			q.sql.grabbed+"("+strings.Join(placeholders, ",")+") FOR UPDATE",
			append([]interface{}{"claimed"}, ids...)...,
		)
		if err != nil {
			log.Debug().Msgf("Failed to select jobs: %s", err)
			return err
		}
		defer rows.Close()

		for i := 0; rows.Next(); i++ {
			var j job
			if err := rows.Scan(&(j.id), &(j.category), &(j.url), &(j.payload), &(j.nextTry), &(j.status), &(j.createdAt), &(j.retryCount), &(j.retryDelay), &(j.failCount), &(j.timeout)); err != nil {
				log.Debug().Msgf("Failed to scan selected jobs: %s", err)
				return err
			}
			j.status = "grabbed"

			ids[i] = j.id
			results = append(results, &j)
		}
		if err := rows.Err(); err != nil {
			log.Debug().Msgf("Failed to read selected jobs: %s", err)
			return err
		}
		return nil
	}(); err != nil {
		tx.Rollback()
		return nil, err
	}

	// The number of jobs may reduce when they are grabbed in another
	// thread right after the preselection.  This is unlikely to
	// happen to a single dispatcher, though.
	if len(results) <= 0 {
		tx.Rollback()
		return results, nil
	}
	placeholders = placeholders[:len(results)]
	ids = ids[:len(results)]

	// 3. UPDATE the status of jobs.
	if err := func() error {
		_, err = tx.Exec(
			q.sql.launch+"("+strings.Join(placeholders, ",")+")",
			ids...,
		)
		if err != nil {
			log.Debug().Msgf("Failed to grab jobs: %s", err)
			return err
		}
		return nil
	}(); err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	// Emulate `ORDER BY next_try ASC`, which causes `using filesort`
	// together with `SELECT ~ WHERE ~ IN`.
	sort.Slice(results, func(i, j int) bool {
		return results[i].(*job).nextTry < results[j].(*job).nextTry
	})

	return results, nil
}

func (q *jobQueue) Delete(completedJob jobqueue.Job) {
	log := q.logger.With().Str("method", "Delete").Logger()

	j, ok := completedJob.(*job)
	if !ok {
		log.Panic().Msgf("Invalid job structure: %v", completedJob)
		return
	}

	if _, err := q.db.Exec(q.sql.deleteJob, j.id); err != nil {
		log.Error().Msgf("Failed to delete a job: %s", err)
	}
}

func (q *jobQueue) Update(completedJob jobqueue.Job, next jobqueue.NextInfo) {
	log := q.logger.With().Str("method", "Update").Logger()

	j, ok := completedJob.(*job)
	if !ok {
		log.Panic().Msgf("Invalid job structure: %v", completedJob)
		return
	}

	if _, err := q.db.Exec(
		q.sql.updateJob,
		next.NextDelay(),
		next.RetryCount(),
		next.FailCount(),
		j.id,
	); err != nil {
		log.Error().Msgf("Failed to update a job: %s", err)
	}
}

func (q *jobQueue) Recover() {
	log := q.logger.With().Str("method", "Recover").Logger()

	q.mu.Lock()
	defer q.mu.Unlock()
	if q.dbPop == nil {
		return
	}

	var recovered int
	placeholders := make([]string, 0)
	ids := make([]interface{}, 0)

	for {
		placeholders = placeholders[:0]
		ids = ids[:0]

		log.Info().Msgf("Recovering orphan jobs...")

		if err := func() error {
			rows, err := q.dbPop.Query(q.sql.orphan)
			if err != nil {
				log.Error().Msgf("Failed to select orphan jobs: %s", err)
				return err
			}
			defer rows.Close()

			for rows.Next() {
				var id uint64
				if err := rows.Scan(&id); err != nil {
					log.Error().Msgf("Failed to scan selected orphan jobs: %s", err)
					return err
				}
				placeholders = append(placeholders, "?")
				ids = append(ids, id)
			}
			if err := rows.Err(); err != nil {
				log.Error().Msgf("Failed to read orphan jobs: %s", err)
				return err
			}
			return nil
		}(); err != nil {
			return
		}
		if len(ids) <= 0 {
			log.Info().Msgf("Recovering complete: %d job(s) recovered", recovered)
			return
		}

		recovered += len(ids)

		if err := func() error {
			if _, err := q.dbPop.Exec(q.sql.recover+"("+strings.Join(placeholders, ",")+")", ids...); err != nil {
				log.Error().Msgf("Failed to recover orphan jobs: %s", err)
				return err
			}
			return nil
		}(); err != nil {
			return
		}
	}
}

func (q *jobQueue) Inspector() jobqueue.Inspector {
	return &inspector{db: q.db, sql: q.sql}
}

func (q *jobQueue) FailureLog() jobqueue.FailureLog {
	return &failureLog{db: q.db, sql: q.sql}
}

func (q *jobQueue) Node() (*jobqueue.Node, error) {
	query := `
		SELECT ID, HOST FROM information_schema.processlist
		WHERE ID = CONNECTION_ID()
	`

	var node jobqueue.Node
	err := q.db.QueryRow(query).Scan(&node.ID, &node.Host)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// Strip the port number
	if i := strings.LastIndex(node.Host, ":"); i >= 0 {
		node.Host = node.Host[0:i]
	}

	return &node, nil
}

func (q *jobQueue) connect() {
	log := q.logger.With().Str("method", "activate").Logger()

	q.mu.Lock()
	defer q.mu.Unlock()

	dbPop, err := sql.Open("mysql", q.dsn)
	if err != nil {
		log.Panic().Msgf("Cannot open DB: %s", err)
	}
	// Restrict connections to prevent connection ID from being changed.
	dbPop.SetMaxOpenConns(1)
	q.dbPop = dbPop
}

func (q *jobQueue) disconnect() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.dbPop != nil {
		q.dbPop.Close()
		q.dbPop = nil
	}
}
