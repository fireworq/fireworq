package jobqueue

import (
	"sync/atomic"
	"time"

	"github.com/paulbellamy/ratecounter"
	"github.com/prometheus/client_golang/prometheus"
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

	totalPushesCounter            prometheus.Counter
	totalPopsCounter              prometheus.Counter
	totalSuccessesCounter         prometheus.Counter
	totalFailuresCounter          prometheus.Counter
	totalPermanentFailuresCounter prometheus.Counter
	totalCompletesCounter         prometheus.Counter
	totalElapsedCounter           prometheus.Counter
}

var (
	totalPushesCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "fireworq", Name: "total_pushes",
		Help: "",
	}, []string{"queue"})
	totalPopsCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "fireworq", Name: "total_pops",
		Help: "",
	}, []string{"queue"})
	totalSuccessesCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "fireworq", Name: "total_successes",
		Help: "",
	}, []string{"queue"})
	totalFailuresCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "fireworq", Name: "total_failures",
		Help: "",
	}, []string{"queue"})
	totalPermanentFailuresCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "fireworq", Name: "total_permanent_failures",
		Help: "",
	}, []string{"queue"})
	totalCompletesCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "fireworq", Name: "total_completes",
		Help: "",
	}, []string{"queue"})
	totalElapsedCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "fireworq", Name: "total_elapsed",
		Help: "",
	}, []string{"queue"})
)

func init() {
	prometheus.MustRegister(totalPushesCounterVec)
	prometheus.MustRegister(totalPopsCounterVec)
	prometheus.MustRegister(totalSuccessesCounterVec)
	prometheus.MustRegister(totalFailuresCounterVec)
	prometheus.MustRegister(totalPermanentFailuresCounterVec)
	prometheus.MustRegister(totalCompletesCounterVec)
	prometheus.MustRegister(totalElapsedCounterVec)
}

func newStats(name string) *stats {
	return &stats{
		totalPushesCounter:            totalPushesCounterVec.With(prometheus.Labels{"queue": name}),
		totalPopsCounter:              totalPopsCounterVec.With(prometheus.Labels{"queue": name}),
		totalSuccessesCounter:         totalSuccessesCounterVec.With(prometheus.Labels{"queue": name}),
		totalFailuresCounter:          totalFailuresCounterVec.With(prometheus.Labels{"queue": name}),
		totalPermanentFailuresCounter: totalPermanentFailuresCounterVec.With(prometheus.Labels{"queue": name}),
		totalCompletesCounter:         totalCompletesCounterVec.With(prometheus.Labels{"queue": name}),
		totalElapsedCounter:           totalElapsedCounterVec.With(prometheus.Labels{"queue": name}),
		pushesPerSecond:               ratecounter.NewRateCounter(1 * time.Second),
		popsPerSecond:                 ratecounter.NewRateCounter(1 * time.Second),
	}
}

func (s *stats) push(num int64) {
	atomic.AddInt64(&s.totalPushes, num)
	s.pushesPerSecond.Incr(num)
	s.totalPushesCounter.Add(float64(num))
}

func (s *stats) pop(num int64) {
	atomic.AddInt64(&s.totalPops, num)
	s.popsPerSecond.Incr(num)
	s.totalPopsCounter.Add(float64(num))
}

func (s *stats) succeed(num int64) {
	atomic.AddInt64(&s.totalSuccesses, num)
	s.totalSuccessesCounter.Add(float64(num))
}

func (s *stats) fail(num int64) {
	atomic.AddInt64(&s.totalFailures, num)
	s.totalFailuresCounter.Add(float64(num))
}

func (s *stats) permanentlyFail(num int64) {
	atomic.AddInt64(&s.totalPermanentFailures, num)
	s.totalPermanentFailuresCounter.Add(float64(num))
}

func (s *stats) complete(num int64) {
	atomic.AddInt64(&s.totalCompletes, num)
	s.totalCompletesCounter.Add(float64(num))
}

func (s *stats) elapsed(t int64) {
	atomic.AddInt64(&s.totalElapsed, t)
	s.totalElapsedCounter.Add(float64(t))
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
