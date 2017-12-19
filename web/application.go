package web

import (
	"io"
	"strconv"
	"time"

	"github.com/fireworq/fireworq/config"
	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/model"
	"github.com/fireworq/fireworq/repository"
	"github.com/fireworq/fireworq/service"

	stats "github.com/fukata/golang-stats-api-handler"
	"github.com/rs/zerolog/log"
)

// Service is an interface of the application use case service.
type Service interface {
	Stop() <-chan struct{}
	GetJobQueue(qn string) (service.RunningQueue, bool)
	DeleteJobQueue(qn string) error
	AddJobQueue(q *model.Queue) error
	Push(job jobqueue.IncomingJob) (*service.PushResult, error)
}

// Application is an interface of the application.
type Application struct {
	AccessLogWriter   io.Writer
	Version           string
	Service           Service
	QueueRepository   repository.QueueRepository
	RoutingRepository repository.RoutingRepository
}

func (app *Application) newServer() *server {
	s := newServer(app.AccessLogWriter)

	s.handle("/", app.serveVersion)
	s.handle("/version", app.serveVersion)
	s.handle("/settings", app.serveSettings)
	s.mux.HandleFunc("/stats", stats.Handler)
	s.handle("/job/{category:.+}", app.serveJob)
	s.handle("/queues", app.serveQueueList)
	s.handle("/queues/stats", app.serveQueueListStats)
	s.handle("/queue/{queue:[^/]+}", app.serveQueue)
	s.handle("/queue/{queue:[^/]+}/node", app.serveQueueNode)
	s.handle("/queue/{queue:[^/]+}/stats", app.serveQueueStats)
	s.handle("/queue/{queue:[^/]+}/grabbed", app.serveQueueGrabbed)
	s.handle("/queue/{queue:[^/]+}/waiting", app.serveQueueWaiting)
	s.handle("/queue/{queue:[^/]+}/deferred", app.serveQueueDeferred)
	s.handle("/queue/{queue:[^/]+}/job/{id:[^/]+}", app.serveQueueJob)
	s.handle("/queue/{queue:[^/]+}/failed", app.serveQueueFailed)
	s.handle("/queue/{queue:[^/]+}/failed/{id:[^/]+}", app.serveQueueFailedJob)
	s.handle("/routings", app.serveRoutingList)
	s.handle("/routing/{category:.+}", app.serveRouting)

	return s
}

// Serve starts the application Web server.
func (app *Application) Serve() {
	shutdownTimeout := time.Duration(shutdownTimeout()) * time.Second
	timeout, err := app.newServer().startGracefully(shutdownTimeout)
	if err != nil {
		log.Warn().Msgf("Stopped the HTTP server: %s", err)
	}

	log.Info().Msgf("Stopping the job dispatcher in %s ...", timeout)
	select {
	case <-time.After(timeout):
		log.Warn().Msg("Stopped the job dispatcher: deadline exceeded")
	case <-app.Service.Stop():
		log.Info().Msg("Stopped the job dispatcher")
	}
}

func shutdownTimeout() uint {
	str := config.Get("shutdown_timeout")
	timeout, err := strconv.Atoi(str)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	return uint(timeout)
}
