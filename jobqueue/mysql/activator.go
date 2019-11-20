package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	activatorLockWaitTimeout    = 10 * time.Second
	activatorActivationInterval = 1 * time.Second
)

type activator struct {
	queueName string
	cancel    atomic.Value
	stoppedC  chan struct{}
	stopped   uint32
	active    int32
	db        *sql.DB
	dsn       string
	logger    zerolog.Logger
}

type activation interface {
	queueName() string
	getDsn() string
}

func startActivator(q activation, onActivating func()) *activator {
	a := &activator{
		queueName: q.queueName(),
		dsn:       q.getDsn(),
		stoppedC:  make(chan struct{}),
		logger:    log.With().Str("queue", q.queueName()).Logger(),
		active:    -1,
	}
	go a.loop(onActivating)

	return a
}

func (a *activator) stop() <-chan struct{} {
	atomic.StoreUint32(&a.stopped, 1)

	if cancel := a.cancel.Load(); cancel != nil {
		cancel.(context.CancelFunc)()
	}

	return a.stoppedC
}

func (a *activator) isActive() bool {
	return atomic.LoadInt32(&a.active) > 0
}

func (a *activator) lockName() string {
	return fmt.Sprintf("fireworq_jq(%s)", a.queueName)
}

// File private methods

func (a *activator) loop(onActivating func()) {
	ticker := time.NewTicker(activatorActivationInterval)
	for a.activate(onActivating) {
		<-ticker.C
	}
	ticker.Stop()

	atomic.StoreInt32(&a.active, 0)
	a.disconnect()
	a.stoppedC <- struct{}{}
}

func (a *activator) activate(onActivating func()) (shouldRetry bool) {
	if atomic.LoadUint32(&a.stopped) > 0 {
		return false
	}

	if err := a.connect(); err != nil {
		// This should not happen since sql.Open() won't try to
		// connect to the DB and won't fail unless the DB driver name
		// is invalid.  We just try again in case sql.Open() changes
		// the behavior in future.
		a.logger.Error().Msgf("(activator) %s", err)
		return true
	}

	if a.hasLock() {
		// Make sure that the queue is active since we have the lock.
		// Without doing this, the queue won't be activated if the
		// former call of getLock() failed to receive packets from the
		// DB but `GET_LOCK` had actually been accepted by the DB.
		if atomic.SwapInt32(&a.active, 1) <= 0 {
			a.logger.Info().Msg("The node is now in PRIMARY mode")
		}
		return true
	}

	if atomic.SwapInt32(&a.active, 0) != 0 {
		a.logger.Info().Msg("The node is now in BACKUP mode")
	}
	a.logger.Debug().Msg("Queue (re)activating...")

	if err := a.getLock(); err != nil {
		if _, ok := err.(*stoppedError); ok {
			return false
		} else if _, ok := err.(*lockTimeoutError); ok {
			// `GET_LOCK` timed out; just try again.
			a.logger.Debug().Msg(err.Error())

			// This doesn't seem to happen to `GET_LOCK`:
			// } else if e, ok := err.(*mysqldriver.MySQLError); ok && e.Number == 1205 {
			// 	// Lock wait timeout (`lock_wait_timeout`) exceeded.
			// 	// Just try again later.
			// 	a.logger.Debug().Msg(err.Error())

		} else {
			// Connection failed (maybe DB server down).
			// Try to reconnect later.
			a.disconnect()
			a.logger.Error().Msgf("(activator) %s", err)
		}

		return true
	}

	a.logger.Info().Msg("Switching to PRIMARY mode...")

	onActivating()
	atomic.StoreInt32(&a.active, 1)

	a.logger.Debug().Msg("Queue activated")
	a.logger.Info().Msg("The node is now in PRIMARY mode")

	return true
}

func (a *activator) connect() error {
	if a.db != nil {
		return nil
	}

	db, err := sql.Open("mysql", a.dsn)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	a.db = db

	return nil
}

func (a *activator) disconnect() {
	if a.db == nil {
		return
	}
	a.db.Close()
	a.db = nil
}

func (a *activator) hasLock() bool {
	query := "SELECT IS_USED_LOCK(?) = CONNECTION_ID()"

	var hasLock bool
	if err := a.db.QueryRow(query, a.lockName()).Scan(&hasLock); err != nil {
		return false
	}

	return hasLock
}

func (a *activator) getLock() error {
	err := func() error {
		ctx, cancel := context.WithCancel(context.Background())
		a.cancel.Store(cancel)

		if atomic.LoadUint32(&a.stopped) > 0 {
			return &stoppedError{}
		}

		var locked sql.NullInt64
		if err := a.db.QueryRowContext(
			ctx,
			"SELECT GET_LOCK(?, ?)",
			a.lockName(),
			int(activatorLockWaitTimeout.Seconds()),
		).Scan(&locked); err != nil {
			return err
		}

		if !locked.Valid {
			// Got NULL; means running out of memory or the thread was killed.
			return &lockError{}
		}

		if locked.Int64 < 1 {
			return &lockTimeoutError{}
		}

		return nil
	}()

	if err == context.Canceled && atomic.LoadUint32(&a.stopped) > 0 {
		return &stoppedError{}
	}
	return err
}

type lockError struct{}

func (err *lockError) Error() string {
	return "GET_LOCK returned NULL"
}

type lockTimeoutError struct{}

func (err *lockTimeoutError) Error() string {
	return "Lock timed out"
}

type stoppedError struct{}

func (err *stoppedError) Error() string {
	return "The activator has already been stopped"
}
