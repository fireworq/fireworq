package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fireworq/fireworq/dispatcher"
	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/model"

	"github.com/gorilla/mux"
)

func (app *Application) serveQueueList(w http.ResponseWriter, req *http.Request) error {
	queues, err := app.QueueRepository.FindAll()
	if err != nil {
		return err
	}

	json, err := json.Marshal(queues)
	if err != nil {
		return err
	}
	writeJSON(w, json)

	return nil
}

func (app *Application) serveQueueListStats(w http.ResponseWriter, req *http.Request) error {
	queues, err := app.QueueRepository.FindAll()
	if err != nil {
		return err
	}

	stats := make(map[string]*Stats)
	for _, q := range queues {
		if queue, ok := app.Service.GetJobQueue(q.Name); ok {
			var activeNodes int64
			if queue.IsActive() {
				activeNodes = 1
			}
			stats[q.Name] = &Stats{
				queue.Stats(),
				queue.WorkerStats(),
				activeNodes,
			}
		}
	}

	j, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	writeJSON(w, j)

	return nil
}

func (app *Application) serveQueue(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	name := vars["queue"]
	var definition model.Queue

	if req.Method == "PUT" {
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&definition)
		if err != nil {
			return errBadRequest.WithDetail(err.Error())
		}
		definition.Name = name

		if err := app.Service.AddJobQueue(&definition); err != nil {
			return err
		}
	} else {
		q, err := app.QueueRepository.FindByName(name)
		if err != nil {
			return errNotFound
		}
		definition = *q

		if req.Method == "DELETE" {
			if err := app.Service.DeleteJobQueue(name); err != nil {
				return err
			}
		}
	}

	j, err := json.Marshal(&definition)
	if err != nil {
		return err
	}

	writeJSON(w, j)
	return nil
}

func (app *Application) serveQueueNode(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	q, ok := app.Service.GetJobQueue(vars["queue"])
	if !ok {
		return errNotFound.WithDetail(fmt.Sprintf("No such queue: %s", vars["queue"]))
	}

	node, err := q.Node()
	if err != nil {
		return err
	}
	if node == nil {
		return errNotFound.WithDetail(fmt.Sprintf("No node is active for this queue"))
	}

	j, err := json.Marshal(node)
	if err != nil {
		return err
	}
	writeJSON(w, j)

	return nil
}

func (app *Application) serveQueueStats(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	q, ok := app.Service.GetJobQueue(vars["queue"])
	if !ok {
		return errNotFound.WithDetail(fmt.Sprintf("No such queue: %s", vars["queue"]))
	}

	var activeNodes int64
	if q.IsActive() {
		activeNodes = 1
	}
	j, err := json.Marshal(&Stats{
		q.Stats(),
		q.WorkerStats(),
		activeNodes,
	})
	if err != nil {
		return err
	}
	writeJSON(w, j)

	return nil
}

func (app *Application) serveQueueGrabbed(w http.ResponseWriter, req *http.Request) error {
	return app.serveQueueJobs(func(i jobqueue.Inspector, l uint, c string) (*jobqueue.InspectedJobs, error) {
		return i.FindAllGrabbed(l, c)
	}, w, req)
}

func (app *Application) serveQueueWaiting(w http.ResponseWriter, req *http.Request) error {
	return app.serveQueueJobs(func(i jobqueue.Inspector, l uint, c string) (*jobqueue.InspectedJobs, error) {
		return i.FindAllWaiting(l, c)
	}, w, req)
}

func (app *Application) serveQueueDeferred(w http.ResponseWriter, req *http.Request) error {
	return app.serveQueueJobs(func(i jobqueue.Inspector, l uint, c string) (*jobqueue.InspectedJobs, error) {
		return i.FindAllDeferred(l, c)
	}, w, req)
}

func (app *Application) serveQueueJobs(find func(jobqueue.Inspector, uint, string) (*jobqueue.InspectedJobs, error), w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	query := req.URL.Query()

	q, ok := app.Service.GetJobQueue(vars["queue"])
	if !ok {
		return errNotFound
	}

	inspector, ok := q.Inspector()
	if !ok {
		return errNotImplemented
	}

	limit := uint(100)
	if l, err := strconv.Atoi(query.Get("limit")); err == nil {
		limit = uint(l)
	}

	jobs, err := find(inspector, limit, query.Get("cursor"))
	if err != nil {
		return err
	}

	j, err := json.Marshal(jobs)
	if err != nil {
		return err
	}
	writeJSON(w, j)

	return nil
}

func (app *Application) serveQueueFailed(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	query := req.URL.Query()

	q, ok := app.Service.GetJobQueue(vars["queue"])
	if !ok {
		return errNotFound
	}

	failureLog, ok := q.FailureLog()
	if !ok {
		return errNotImplemented
	}

	var findAll func(uint, string) (*jobqueue.FailedJobs, error)
	if query.Get("order") == "created" {
		findAll = failureLog.FindAll
	} else {
		findAll = failureLog.FindAllRecentFailures
	}

	limit := uint(100)
	if l, err := strconv.Atoi(query.Get("limit")); err == nil {
		limit = uint(l)
	}

	jobs, err := findAll(limit, query.Get("cursor"))
	if err != nil {
		return err
	}

	j, err := json.Marshal(jobs)
	if err != nil {
		return err
	}
	writeJSON(w, j)

	return nil
}

func (app *Application) serveQueueJob(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errBadRequest
	}

	q, ok := app.Service.GetJobQueue(vars["queue"])
	if !ok {
		return errNotFound
	}

	inspector, ok := q.Inspector()
	if !ok {
		return errNotImplemented
	}

	job, err := inspector.Find(uint64(id))
	if err == sql.ErrNoRows {
		return errNotFound
	}
	if err != nil {
		return err
	}

	if req.Method == "DELETE" {
		if err := inspector.Delete(uint64(id)); err != nil {
			return err
		}
	}

	j, err := json.Marshal(job)
	if err != nil {
		return err
	}
	writeJSON(w, j)

	return nil
}

func (app *Application) serveQueueFailedJob(w http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return errBadRequest
	}

	q, ok := app.Service.GetJobQueue(vars["queue"])
	if !ok {
		return errNotFound
	}

	failureLog, ok := q.FailureLog()
	if !ok {
		return errNotImplemented
	}

	job, err := failureLog.Find(uint64(id))
	if err == sql.ErrNoRows {
		return errNotFound
	}
	if err != nil {
		return err
	}

	if req.Method == "DELETE" {
		if err := failureLog.Delete(uint64(id)); err != nil {
			return err
		}
	}

	j, err := json.Marshal(job)
	if err != nil {
		return err
	}
	writeJSON(w, j)

	return nil
}

// JobqueueStats is an alias to pointer type of jobqueue.Stats.
type JobqueueStats = *jobqueue.Stats

// DispatcherStats is an alias to pointer type of dispatcher.Stats.
type DispatcherStats = *dispatcher.Stats

// Stats contains queue statistics and worker statistics.
type Stats struct {
	JobqueueStats
	DispatcherStats
	ActiveNodes int64 `json:"active_nodes"`
}
