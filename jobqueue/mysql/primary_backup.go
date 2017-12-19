package mysql

import (
	"database/sql"
	"strings"

	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/model"
)

type primaryBackupJobQueue struct {
	*jobQueue
	activator *activator
}

// NewPrimaryBackup creates a jobqueue.Impl which uses MySQL as a data
// store and restricts only one node to be active in a cluster.
//
// Inactive nodes become backup nodes, which will be active when the
// active node dies.
func NewPrimaryBackup(definition *model.Queue, dsn string) jobqueue.Impl {
	q := newJobQueue(definition, dsn)
	return &primaryBackupJobQueue{q, nil}
}

func (q *primaryBackupJobQueue) Start() {
	q.jobQueue.Start()
	q.activator = startActivator(
		q,
		q.Recover,
	)
}

func (q *primaryBackupJobQueue) Stop() <-chan struct{} {
	stopped := make(chan struct{})
	go func() {
		<-q.activator.stop()
		<-q.jobQueue.Stop()
		stopped <- struct{}{}
	}()
	return stopped
}

func (q *primaryBackupJobQueue) IsActive() bool {
	return q.activator.isActive()
}

func (q *primaryBackupJobQueue) Pop(limit uint) ([]jobqueue.Job, error) {
	if !q.IsActive() {
		return nil, &jobqueue.InactiveError{}
	}

	return q.jobQueue.Pop(limit)
}

func (q *primaryBackupJobQueue) Node() (*jobqueue.Node, error) {
	query := `
		SELECT ID, HOST FROM information_schema.processlist
		WHERE ID = IS_USED_LOCK(?)
	`

	var node jobqueue.Node
	err := q.db.QueryRow(query, q.activator.lockName()).Scan(&node.ID, &node.Host)
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

// activation interface

func (q *primaryBackupJobQueue) queueName() string {
	return q.name
}

func (q *primaryBackupJobQueue) getDsn() string {
	return q.dsn
}
