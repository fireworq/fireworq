package logger

import (
	"os"
	"time"

	"github.com/fireworq/fireworq/config"
	logwriter "github.com/fireworq/fireworq/log"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// Writer is a log writer which logger.Info() or logger.Debug()
	// writes to.
	Writer = logwriter.New(os.Stdout)

	tag    string
	logger *zerolog.Logger = func() *zerolog.Logger {
		l := zerolog.Nop()
		return &l
	}()
)

// Init initializes global parameters of logger by configuration values.
//
// Configuration keys prefixed by "queue_log_" are considered.
func Init() {
	tag = config.Get("queue_log_tag")

	output := config.Get("queue_log")
	if len(output) > 0 {
		log.Info().Msgf("Queue log file: %s", output)
		w, err := logwriter.OpenFile(output)
		if err != nil {
			log.Panic().Msg(err.Error())
		}
		Writer = w
	}

	level := logwriter.ParseLevel(config.Get("queue_log_level"), zerolog.InfoLevel)

	l := zerolog.New(Writer).With().Logger().Level(level)
	logger = &l
}

func put(event *zerolog.Event, queue string, action string, j LoggableJob, msg string) {
	created := int64(j.CreatedAt())
	elapsed := Elapsed(j)
	event.
		Int64("time", created+elapsed).
		Str("tag", tag).
		Str("action", action).
		Str("queue", queue).
		Str("category", j.Category()).
		Uint64("id", j.ID()).
		Str("status", j.Status()).
		Int64("created_at", created).
		Int64("elapsed", elapsed).
		Str("url", j.URL()).
		Str("payload", j.Payload()).
		Uint64("next_try", j.NextTry()).
		Uint("retry_count", j.RetryCount()).
		Uint("retry_delay", j.RetryDelay()).
		Uint("fail_count", j.FailCount()).
		Uint("timeout", j.Timeout()).
		Msg(msg)
}

// Info writes an INFO level log entry of a job action.
func Info(queue string, action string, j LoggableJob, msg string) {
	put(logger.Info(), queue, action, j, msg)
}

// Debug writes a DEBUG level log entry of a job action.
func Debug(queue string, action string, j LoggableJob, msg string) {
	put(logger.Debug(), queue, action, j, msg)
}

// LoggableJob defines fields of a job to be written into the log.
type LoggableJob interface {
	Category() string
	URL() string
	Payload() string

	ID() uint64
	Status() string

	NextTry() uint64
	RetryCount() uint
	RetryDelay() uint
	FailCount() uint
	Timeout() uint

	CreatedAt() uint64
}

// Elapsed returns elapsed since the job is created in millisecond.
func Elapsed(j LoggableJob) int64 {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	created := int64(j.CreatedAt())
	return now - created
}
