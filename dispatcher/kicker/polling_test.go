package kicker

import (
	"sync/atomic"
	"testing"
	"time"
)

type dummyKickable struct {
	kicked int64
}

func (k *dummyKickable) Kick() {
	atomic.AddInt64(&k.kicked, 1)
}

func TestStartStop(t *testing.T) {
	cfg := PollingKicker{Interval: uint(100)}
	k := cfg.NewKicker()

	k.Start(&dummyKickable{})

	select {
	case <-k.Stop():
	case <-time.After(3 * time.Second):
		t.Error("A polling kicker should be able to stop")
	}
}

func TestPollingInterval(t *testing.T) {
	interval := uint(123)
	cfg := PollingKicker{Interval: interval}
	k := cfg.NewKicker()

	if k.PollingInterval() != interval {
		t.Error("A polling kicker should return its interval")
	}
}

func TestPing(t *testing.T) {
	cfg := PollingKicker{Interval: uint(100)}
	k := cfg.NewKicker()

	k.Start(&dummyKickable{})
	defer func() { <-k.Stop() }()

	k.Ping()
}

func TestLoop(t *testing.T) {
	cfg := PollingKicker{Interval: uint(100)}
	k := cfg.NewKicker()
	kickable := &dummyKickable{}
	k.Start(kickable)

	<-time.After(1 * time.Second)
	if atomic.LoadInt64(&kickable.kicked) < 1 {
		t.Error("A polling kicker should kick")
	}
	<-k.Stop()
}
