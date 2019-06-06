package jobqueue

import (
	"sync/atomic"
	"time"

	"github.com/paulbellamy/ratecounter"
)

// Stats describes queue statistics.
type Stats struct {
	TotalPushes            int64 `json:"total_pushes"`
	TotalPops              int64 `json:"total_pops"`
	TotalSuccesses         int64 `json:"total_successes"`
	TotalFailures          int64 `json:"total_failures"`
	TotalPermanentFailures int64 `json:"total_permanent_failures"`
	TotalCompletes         int64 `json:"total_completes"`
	TotalElapsed           int64 `json:"total_elapsed"`
	PushesPerSecond        int64 `json:"pushes_per_second"`
	PopsPerSecond          int64 `json:"pops_per_second"`
}

type stats struct {
	totalPushes            int64
	totalPops              int64
	totalSuccesses         int64
	totalFailures          int64
	totalPermanentFailures int64
	totalCompletes         int64
	totalElapsed           int64
	pushesPerSecond        *ratecounter.RateCounter
	popsPerSecond          *ratecounter.RateCounter
}

func newStats() *stats {
	return &stats{
		pushesPerSecond: ratecounter.NewRateCounter(1 * time.Second),
		popsPerSecond:   ratecounter.NewRateCounter(1 * time.Second),
	}
}

func (s *stats) push(num int64) {
	atomic.AddInt64(&s.totalPushes, num)
	s.pushesPerSecond.Incr(num)
}

func (s *stats) pop(num int64) {
	atomic.AddInt64(&s.totalPops, num)
	s.popsPerSecond.Incr(num)
}

func (s *stats) succeed(num int64) {
	atomic.AddInt64(&s.totalSuccesses, num)
}

func (s *stats) fail(num int64) {
	atomic.AddInt64(&s.totalFailures, num)
}

func (s *stats) permanentlyFail(num int64) {
	atomic.AddInt64(&s.totalPermanentFailures, num)
}

func (s *stats) complete(num int64) {
	atomic.AddInt64(&s.totalCompletes, num)
}

func (s *stats) elapsed(t int64) {
	atomic.AddInt64(&s.totalElapsed, t)
}

func (s *stats) export() *Stats {
	return &Stats{
		TotalPushes:            atomic.LoadInt64(&s.totalPushes),
		TotalPops:              atomic.LoadInt64(&s.totalPops),
		TotalSuccesses:         atomic.LoadInt64(&s.totalSuccesses),
		TotalFailures:          atomic.LoadInt64(&s.totalFailures),
		TotalPermanentFailures: atomic.LoadInt64(&s.totalPermanentFailures),
		TotalCompletes:         atomic.LoadInt64(&s.totalCompletes),
		TotalElapsed:           atomic.LoadInt64(&s.totalElapsed),
		PushesPerSecond:        s.pushesPerSecond.Rate(),
		PopsPerSecond:          s.popsPerSecond.Rate(),
	}
}
