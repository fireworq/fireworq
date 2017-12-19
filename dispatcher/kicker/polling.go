package kicker

import (
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// PollingKicker is a builder of a Kicker which kicks a Kickable
// repeatedly on some interval.
type PollingKicker struct {
	Interval uint
}

// NewKicker creates a new polling kicker instance.
func (cfg *PollingKicker) NewKicker() Kicker {
	log.Debug().Msgf("Polling interval: %d", cfg.Interval)
	return &pollingKicker{
		interval: cfg.Interval,
		stop:     make(chan struct{}, 1),
		stopped:  make(chan struct{}, 1),
	}
}

type pollingKicker struct {
	interval uint
	started  uint32
	stop     chan struct{}
	stopped  chan struct{}
}

func (k *pollingKicker) Start(kickable Kickable) {
	go k.loop(kickable)
}

func (k *pollingKicker) Stop() <-chan struct{} {
	k.stop <- struct{}{}
	return k.stopped
}

func (k *pollingKicker) Ping() {
	// ignore; do nothing
}

func (k *pollingKicker) PollingInterval() uint {
	return k.interval
}

func (k *pollingKicker) loop(kickable Kickable) {
	ticker := time.NewTicker(time.Duration(k.interval) * time.Millisecond)
Loop:
	for {
		select {
		case <-ticker.C:
			kickable.Kick()
		case <-k.stop:
			ticker.Stop()
			atomic.StoreUint32(&k.started, 0)
			break Loop
		}
	}
	k.stopped <- struct{}{}
}
