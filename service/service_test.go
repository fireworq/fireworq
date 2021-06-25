package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fireworq/fireworq/config"
	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/model"
	repository "github.com/fireworq/fireworq/repository/factory"
	"github.com/fireworq/fireworq/test"
)

func TestMain(m *testing.M) {
	test.RunAll(m)
}

func TestAddJobQueue(t *testing.T) {
	queueName := "service_create_test_queue"

	svc := newService()
	defer func() { <-svc.Stop() }()
	defer svc.DeleteJobQueue(queueName)

	func() {
		q, ok := svc.GetJobQueue(queueName)
		if q != nil || ok {
			t.Error("Undefined queue should not be retrieved")
		}
	}()

	func() {
		var pollingInterval uint = 300
		var maxWorkers uint = 10
		q := &model.Queue{
			Name:            queueName,
			PollingInterval: pollingInterval,
			MaxWorkers:      maxWorkers,
		}
		err := svc.AddJobQueue(q)
		if err != nil {
			t.Error(err)
		}

		q1, ok := svc.GetJobQueue(queueName)
		if q1 == nil || !ok {
			t.Error("Defined queue should be retrieved")
		}

		if q1.Name() != queueName || q1.PollingInterval() != pollingInterval || q1.MaxWorkers() != maxWorkers {
			t.Error("Defined queue should store the defined value")
		}
	}()

	queueName = "service_create_test_queue1"
	defer svc.DeleteJobQueue(queueName)

	func() {
		q := &model.Queue{Name: queueName}
		err := svc.AddJobQueue(q)
		if err != nil {
			t.Error(err)
		}

		q1, ok := svc.GetJobQueue(queueName)
		if q1 == nil || !ok {
			t.Error("Defined queue should be retrieved")
		}

		if q1.Name() != queueName {
			t.Error("Defined queue should store the defined value")
		}
		if q1.PollingInterval() == 0 || q1.MaxWorkers() == 0 {
			t.Error("Defined queue should have default values")
		}
	}()

	func() {
		var pollingInterval uint = 300
		var maxWorkers uint = 10
		q := &model.Queue{
			Name:            queueName,
			PollingInterval: pollingInterval,
			MaxWorkers:      maxWorkers,
		}
		err := svc.AddJobQueue(q)
		if err != nil {
			t.Error(err)
		}

		q1, ok := svc.GetJobQueue(queueName)
		if q1 == nil || !ok {
			t.Error("Defined queue should be retrieved")
		}

		if q1.Name() != queueName {
			t.Error("Defined queue should store the defined value")
		}
		if q1.PollingInterval() != pollingInterval || q1.MaxWorkers() != maxWorkers {
			t.Error("Defined queue should store the overridden values")
		}
	}()

	func() {
		var pollingInterval uint = 300
		var maxWorkers uint = 10
		var maxDispatchesPerSecond float64 = 2.5
		var maxBurstSize uint = 5
		q := &model.Queue{
			Name:                   queueName,
			PollingInterval:        pollingInterval,
			MaxWorkers:             maxWorkers,
			MaxDispatchesPerSecond: maxDispatchesPerSecond,
			MaxBurstSize:           maxBurstSize,
		}
		err := svc.AddJobQueue(q)
		if err != nil {
			t.Error(err)
		}

		q1, ok := svc.GetJobQueue(queueName)
		if q1 == nil || !ok {
			t.Error("Defined queue should be retrieved")
		}

		if q1.Name() != queueName {
			t.Error("Defined queue should store the defined value")
		}
		if q1.PollingInterval() != throttleQueuePollingInterval {
			t.Error("Defined queue should store the fixed polling interval")
		}
		if q1.MaxWorkers() != maxWorkers {
			t.Error("Defined queue should store the overridden values")
		}
	}()

	func() {
		q := &model.Queue{
			Name:                   queueName,
			MaxDispatchesPerSecond: -0.1,
		}
		err := svc.AddJobQueue(q)
		if err == nil {
			t.Error("AddJobQueue should fail with negative MaxDispatchesPerSecond")
		}
	}()

	func() {
		q := &model.Queue{
			Name:         queueName,
			MaxBurstSize: 5,
		}
		err := svc.AddJobQueue(q)
		if err == nil {
			t.Error("AddJobQueue should fail with MaxBurstSize but without MaxDispatchesPerSecond")
		}
	}()
}

func TestDeleteJobQueue(t *testing.T) {
	svc := newService()
	defer func() { <-svc.Stop() }()

	name := "service_delete_test_queue"

	func() {
		var pollingInterval uint = 300
		var maxWorkers uint = 10
		q := &model.Queue{
			Name:            name,
			PollingInterval: pollingInterval,
			MaxWorkers:      maxWorkers,
		}
		err := svc.AddJobQueue(q)
		if err != nil {
			t.Error(err)
		}

		q1, ok := svc.GetJobQueue(name)
		if q1 == nil || !ok {
			t.Error("Defined queue should be retrieved")
		}

		if q1.Name() != name || q1.PollingInterval() != pollingInterval || q1.MaxWorkers() != maxWorkers {
			t.Error("Defined queue should store the defined value")
		}

		svc.DeleteJobQueue(name)

		q2, ok := svc.GetJobQueue(name)
		if q2 != nil || ok {
			t.Error("Deleted queue should not be retrieved")
		}
	}()
}

func TestDefaultQueue(t *testing.T) {
	queueName := "service_test_default_queue"
	config.Locally("queue_default", queueName, func() {
		svc := newService()
		defer func() { <-svc.Stop() }()

		q, ok := svc.GetJobQueue(queueName)
		if q == nil || !ok {
			t.Error("The default queue should be retrieved")
		}

		if q.Name() != queueName {
			t.Error("Wrong queue name")
		}

	})

	r := repository.NewRepositories()
	r.Queue.DeleteByName(queueName)
}

func TestBindFailure(t *testing.T) {
	if test.If("driver", "in-memory") { // not supported
		return
	}

	svc := newService()
	defer func() { <-svc.Stop() }()

	jobCategory := "service_bind_failure_test_job"
	queueName := "service_bind_failure_test_queue"

	if svc.routing.Add(jobCategory, queueName) == nil {
		t.Error("Binding undefined queue to a job should fail")
	}
}

func TestPush(t *testing.T) {
	jobCategory := "service_push_test_job"
	queueName := "service_push_test_queue"

	svc := newService()
	defer func() { <-svc.Stop() }()
	defer svc.DeleteJobQueue(queueName)

	func() {
		q := &model.Queue{Name: queueName, MaxWorkers: uint(10)}
		err := svc.AddJobQueue(q)
		if err != nil {
			t.Error(err)
		}
	}()

	if err := svc.routing.Add(jobCategory, queueName); err != nil {
		t.Error(err)
	}

	time.Sleep(100 * time.Millisecond) // wait for up

	worker := newTestWorker(t)
	defer worker.close()

	job1 := &incomingJob{
		category: jobCategory,
		url:      worker.url(),
		payload:  "foo bar",
	}

	job2 := &incomingJob{
		category: jobCategory,
		url:      worker.url(),
		payload:  "baz qux",
	}

	r, err := svc.Push(job1)
	if err != nil {
		t.Error(err)
	}
	if worker.wait(3*time.Second) != "foo bar" {
		t.Error("Pushed job should be fired")
	}
	if r.QueueName != "service_push_test_queue" {
		t.Error("Job must be push to a routed queue")
	}

	if _, err := svc.Push(job2); err != nil {
		t.Error(err)
	}
	if worker.wait(3*time.Second) != "baz qux" {
		t.Error("Pushed job should be fired")
	}
}

func TestPushFailure(t *testing.T) {
	svc := newService()
	defer func() { <-svc.Stop() }()

	jobCategory := "service_push_failure_test_job"

	job1 := &incomingJob{
		category: jobCategory,
		url:      "http://localhost/",
		payload:  "foo bar",
	}

	if _, err := svc.Push(job1); err == nil {
		t.Error("Pushing a job before defining its routing should fail")
	}
}

func TestFailingOver(t *testing.T) {
	if test.If("driver", "in-memory") { // not supported
		return
	}

	jobCategory := "service_failing_over_test_job"
	queueName := "service_failing_over_test_queue"

	svc1 := newService()

	func() {
		q := &model.Queue{Name: queueName, MaxWorkers: uint(10)}
		err := svc1.AddJobQueue(q)
		if err != nil {
			t.Error(err)
		}
	}()

	if err := svc1.routing.Add(jobCategory, queueName); err != nil {
		t.Error(err)
	}

	worker := newTestWorker(t)
	defer worker.close()

	time.Sleep(100 * time.Millisecond) // wait for up

	// prepare backup service
	svc2 := newService()

	job1 := &incomingJob{
		category: jobCategory,
		url:      worker.url(),
		payload:  "foo bar",
	}

	job2 := &incomingJob{
		category: jobCategory,
		url:      worker.url(),
		payload:  "baz qux",
	}

	job3 := &incomingJob{
		category: jobCategory,
		url:      worker.url(),
		payload:  "foo qux",
	}

	if _, err := svc1.Push(job1); err != nil {
		t.Error(err)
	}
	if worker.wait(3*time.Second) != "foo bar" {
		t.Error("Pushed job should be fired")
	}

	if _, err := svc2.Push(job2); err != nil {
		t.Error(err)
	}
	if worker.wait(3*time.Second) != "baz qux" {
		t.Error("Pushed job should be fired")
	}

	<-svc1.Stop()
	time.Sleep(100 * time.Millisecond) // wait for up

	if _, err := svc2.Push(job3); err != nil {
		t.Error(err)
	}
	if worker.wait(3*time.Second) != "foo qux" {
		t.Error("Pushed job should be fired")
	}

	svc2.DeleteJobQueue(queueName)
	<-svc2.Stop()
}

func TestRoutingMishit(t *testing.T) {
	jobCategory := "service_routing_mishit_test_job"
	queueName := "service_routing_mishit_test_queue"

	svc := newService()
	defer func() { <-svc.Stop() }()
	defer svc.DeleteJobQueue(queueName)

	func() {
		q := &model.Queue{Name: queueName, MaxWorkers: uint(10)}
		err := svc.AddJobQueue(q)
		if err != nil {
			t.Error(err)
		}
	}()

	time.Sleep(100 * time.Millisecond) // wait for up

	worker := newTestWorker(t)
	defer worker.close()

	job := &incomingJob{
		category: jobCategory,
		url:      worker.url(),
		payload:  "foo bar",
	}

	if _, err := svc.Push(job); err == nil {
		t.Error("Pushed job should not be delivered if it has no routing")
	}

	if err := svc.routing.Add(jobCategory, queueName); err != nil {
		t.Error(err)
	}

	if _, err := svc.Push(job); err != nil {
		t.Error(err)
	}
	if worker.wait(5*time.Second) != "foo bar" {
		t.Error("Pushed job should be fired")
	}

	if err := svc.routing.DeleteByJobCategory(jobCategory); err != nil {
		t.Error(err)
	}
	if _, err := svc.Push(job); err == nil {
		t.Error("Pushed job should not be delivered if it has no routing")
	}
}

func TestReloadingRoutings(t *testing.T) {
	if test.If("driver", "in-memory") { // not supported
		return
	}

	config.Locally("config_refresh_interval", "10", func() {
		jobCategory := "service_reloading_routing_test_job"
		queueName := "service_reloading_routing_test_queue"

		svc1 := newService()
		defer func() { <-svc1.Stop() }()
		defer svc1.DeleteJobQueue(queueName)
		svc2 := newService()
		defer func() { <-svc2.Stop() }()

		func() {
			q := &model.Queue{Name: queueName, MaxWorkers: uint(10)}
			err := svc1.AddJobQueue(q)
			if err != nil {
				t.Error(err)
			}
		}()

		if err := svc1.routing.Add(jobCategory, queueName); err != nil {
			t.Error(err)
		}

		time.Sleep(100 * time.Millisecond) // wait for up

		worker := newTestWorker(t)
		defer worker.close()

		job1 := &incomingJob{
			category: jobCategory,
			url:      worker.url(),
			payload:  "foo bar",
		}

		job2 := &incomingJob{
			category: jobCategory,
			url:      worker.url(),
			payload:  "baz qux",
		}

		if _, err := svc1.Push(job1); err != nil {
			t.Error(err)
		}
		if worker.wait(3*time.Second) != "foo bar" {
			t.Error("Pushed job should be fired")
		}

		time.Sleep(100 * time.Millisecond)

		if _, err := svc2.Push(job2); err != nil {
			t.Error(err)
		}
		if worker.wait(3*time.Second) != "baz qux" {
			t.Error("Pushed job should be fired")
		}

		if err := svc1.routing.DeleteByJobCategory(jobCategory); err != nil {
			t.Error(err)
		}
		if _, err := svc1.Push(job1); err == nil {
			t.Error("Pushed job should not be delivered if it has no routing")
		}

		time.Sleep(100 * time.Millisecond)

		if _, err := svc2.Push(job1); err == nil {
			t.Error("Pushed job should not be delivered if it has no routing")
		}
	})
}

func TestReloadingQueueDefinitions(t *testing.T) {
	if test.If("driver", "in-memory") { // not supported
		return
	}

	config.Locally("config_refresh_interval", "10", func() {
		queueName := "service_reloading_test_queue1"

		svc1 := newService()
		defer func() { <-svc1.Stop() }()
		defer svc1.DeleteJobQueue(queueName)
		svc2 := newService()
		defer func() { <-svc2.Stop() }()

		func() {
			workers := uint(10)
			q := &model.Queue{Name: queueName, MaxWorkers: workers}
			if err := svc1.AddJobQueue(q); err != nil {
				t.Error(err)
			}
			time.Sleep(100 * time.Millisecond) // wait for up

			jq, ok := svc2.GetJobQueue(queueName)
			if !ok || jq.MaxWorkers() != workers {
				t.Error("A service should reload definitions modified by another service instance")
			}
		}()

		func() {
			workers := uint(20)
			q := &model.Queue{Name: queueName, MaxWorkers: workers}
			if err := svc1.AddJobQueue(q); err != nil {
				t.Error(err)
			}
			time.Sleep(100 * time.Millisecond) // wait for up

			jq, ok := svc2.GetJobQueue(queueName)
			if !ok || jq.MaxWorkers() != workers {
				t.Error("A service should reload definitions modified by another service instance")
			}
		}()

		func() {
			if err := svc1.DeleteJobQueue(queueName); err != nil {
				t.Error(err)
			}
			time.Sleep(100 * time.Millisecond)

			jq, ok := svc2.GetJobQueue(queueName)
			if jq != nil || ok {
				t.Error("A deleted queue should not be retrieved")
			}
		}()
	})

	config.Locally("config_refresh_interval", "100000", func() {
		queueName := "service_reloading_test_queue2"
		jobCategory := "service_reloading_test_queue2_job"

		svc1 := newService()
		defer func() { <-svc1.Stop() }()
		defer svc1.DeleteJobQueue(queueName)
		svc2 := newService()
		defer func() { <-svc2.Stop() }()

		workers := uint(10)
		q := &model.Queue{Name: queueName, MaxWorkers: workers}
		if err := svc1.AddJobQueue(q); err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond) // wait for up

		if err := svc1.routing.Add(jobCategory, queueName); err != nil {
			t.Error(err)
		}

		worker := newTestWorker(t)
		defer worker.close()

		job := &incomingJob{
			category: jobCategory,
			url:      worker.url(),
			payload:  `{"status": "success"}`,
		}
		if _, err := svc2.Push(job); err != nil {
			t.Error(err)
		}
		worker.wait(3 * time.Second)
	})

	config.Locally("config_refresh_interval", "100000", func() {
		queueName := "service_reloading_test_queue3"
		jobCategory := "service_reloading_test_queue3_job"

		svc1 := newService()
		defer func() { <-svc1.Stop() }()
		defer svc1.DeleteJobQueue(queueName)
		svc2 := newService()
		defer func() { <-svc2.Stop() }()

		workers := uint(10)
		q := &model.Queue{Name: queueName, MaxWorkers: workers}
		if err := svc1.AddJobQueue(q); err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond) // wait for up

		if err := svc2.routing.Add(jobCategory, queueName); err != nil {
			t.Error(err)
		}

		worker := newTestWorker(t)
		defer worker.close()

		job := &incomingJob{
			category: jobCategory,
			url:      worker.url(),
			payload:  `{"status": "success"}`,
		}
		if _, err := svc2.Push(job); err != nil {
			t.Error(err)
		}
		worker.wait(3 * time.Second)
	})
}

func TestFailureLogging(t *testing.T) {
	jobCategory := "service_failure_log_test_job"
	queueName := "service_failure_log_test_queue"

	svc := newService()
	defer func() { <-svc.Stop() }()
	defer svc.DeleteJobQueue(queueName)

	func() {
		q := &model.Queue{Name: queueName, MaxWorkers: uint(10)}
		err := svc.AddJobQueue(q)
		if err != nil {
			t.Error(err)
		}
	}()

	func() {
		err := svc.routing.Add(jobCategory, queueName)
		if err != nil {
			t.Error(err)
		}
	}()

	worker := newTestWorker(t)
	defer worker.close()

	waitRequest := func() { worker.wait(10 * time.Second) }

	job0 := &incomingJob{
		category: jobCategory,
		url:      worker.url(),
		payload:  `{"status": "failure", "message":"job0"}`,
	}

	job1 := &incomingJob{
		category: jobCategory,
		url:      worker.url(),
		payload:  `$foo`,
	}

	job2 := &incomingJob{
		category:   jobCategory,
		url:        worker.url(),
		payload:    `{}`,
		retryCount: 4,
		retryDelay: 1,
	}

	job3 := &incomingJob{
		category:   jobCategory,
		url:        worker.url(),
		payload:    `{"status": "permanent-failure", "message":"job3"}`,
		retryCount: 3,
		retryDelay: 1,
	}

	jobX := &incomingJob{
		category: jobCategory,
		url:      worker.url(),
		payload:  `{"status": "success", "message":"jobX"}`,
	}

	job4 := &incomingJob{
		category: jobCategory,
		url:      worker.url(),
		payload:  `{"status": "failure", "message":"job4"}`,
	}

	if _, err := svc.Push(job0); err != nil {
		t.Error(err)
	}
	waitRequest()

	if _, err := svc.Push(job1); err != nil {
		t.Error(err)
	}
	waitRequest()

	if _, err := svc.Push(job2); err != nil {
		t.Error(err)
	}
	waitRequest()

	if _, err := svc.Push(job3); err != nil {
		t.Error(err)
	}
	waitRequest()
	waitRequest()

	if _, err := svc.Push(jobX); err != nil {
		t.Error(err)
	}
	waitRequest()
	waitRequest()

	if _, err := svc.Push(job4); err != nil {
		t.Error(err)
	}
	waitRequest()
	waitRequest()

	waitRequest()

	time.Sleep(500 * time.Millisecond) // wait for dispatcher to complete

	q, ok := svc.GetJobQueue(queueName)
	if !ok {
		t.Error("Cannot get queue instance")
	}
	l, ok := q.FailureLog()
	if test.If("driver", "in-memory") {
		if l != nil || ok {
			t.Error("Implemented?")
		}
		return
	}
	if !ok {
		t.Error("Cannot get the failure log")
	}

	func() {
		r, err := l.FindAll(4, "")
		if err != nil {
			t.Error(err)
		}
		jobs := r.FailedJobs
		if len(jobs) != 4 {
			t.Errorf("The number of failed jobs is incorrect: %d", len(jobs))
		}
		if jobs[0].Result.Message != "job4" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if jobs[1].Result.Message != "job3" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if !strings.HasPrefix(jobs[2].Result.Message, "Invalid result status") {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if !strings.HasPrefix(jobs[3].Result.Message, "Cannot parse body as JSON") {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if string(jobs[3].Payload) != `"`+job1.payload+`"` {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if _, err := json.Marshal(r); err != nil {
			t.Error(err)
		}

		for _, j := range jobs {
			job, err := l.Find(j.ID)
			if err != nil {
				t.Error(err)
			}
			if job.Result.Message != j.Result.Message {
				t.Errorf("Wrong job: %v", job)
			}
		}
	}()
	func() {
		r, err := l.FindAll(3, "")
		if err != nil {
			t.Error(err)
		}
		jobs := r.FailedJobs
		r, err = l.FindAll(3, r.NextCursor)
		if err != nil {
			t.Error(err)
		}
		jobs = append(jobs, r.FailedJobs...)
		if len(jobs) != 5 {
			t.Errorf("The number of failed jobs is incorrect: %d", len(jobs))
		}
		if jobs[4].Result.Message != "job0" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if _, err := json.Marshal(r); err != nil {
			t.Error(err)
		}

		for _, j := range jobs {
			job, err := l.Find(j.ID)
			if err != nil {
				t.Error(err)
			}
			if job.Result.Message != j.Result.Message {
				t.Errorf("Wrong job: %v", job)
			}
		}
	}()
	func() {
		r, err := l.FindAll(10, "")
		if err != nil {
			t.Error(err)
		}
		jobs := r.FailedJobs
		if len(jobs) != 5 {
			t.Errorf("The number of failed jobs is incorrect: %d", len(jobs))
		}
		if jobs[4].Result.Message != "job0" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if _, err := json.Marshal(r); err != nil {
			t.Error(err)
		}

		for _, j := range jobs {
			job, err := l.Find(j.ID)
			if err != nil {
				t.Error(err)
			}
			if job.Result.Message != j.Result.Message {
				t.Errorf("Wrong job: %v", job)
			}
		}
	}()

	func() {
		r, err := l.FindAllRecentFailures(4, "")
		if err != nil {
			t.Error(err)
		}
		jobs := r.FailedJobs
		if len(jobs) != 4 {
			t.Errorf("The number of failed jobs is incorrect: %d", len(jobs))
		}
		if !strings.HasPrefix(jobs[0].Result.Message, "Invalid result status") {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if jobs[1].Result.Message != "job4" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if jobs[2].Result.Message != "job3" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if !strings.HasPrefix(jobs[3].Result.Message, "Cannot parse body as JSON") {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if string(jobs[3].Payload) != `"`+job1.payload+`"` {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if _, err := json.Marshal(r); err != nil {
			t.Error(err)
		}

		for _, j := range jobs {
			job, err := l.Find(j.ID)
			if err != nil {
				t.Error(err)
			}
			if job.Result.Message != j.Result.Message {
				t.Errorf("Wrong job: %v", job)
			}
		}
	}()
	func() {
		r, err := l.FindAllRecentFailures(3, "")
		if err != nil {
			t.Error(err)
		}
		jobs := r.FailedJobs
		r, err = l.FindAllRecentFailures(3, r.NextCursor)
		if err != nil {
			t.Error(err)
		}
		jobs = append(jobs, r.FailedJobs...)
		if len(jobs) != 5 {
			t.Errorf("The number of failed jobs is incorrect: %d", len(jobs))
		}
		if jobs[4].Result.Message != "job0" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if _, err := json.Marshal(r); err != nil {
			t.Error(err)
		}

		for _, j := range jobs {
			job, err := l.Find(j.ID)
			if err != nil {
				t.Error(err)
			}
			if job.Result.Message != j.Result.Message {
				t.Errorf("Wrong job: %v", job)
			}
		}
	}()
	func() {
		r, err := l.FindAllRecentFailures(10, "")
		if err != nil {
			t.Error(err)
		}
		jobs := r.FailedJobs
		if len(jobs) != 5 {
			t.Errorf("The number of failed jobs is incorrect: %d", len(jobs))
		}
		if jobs[4].Result.Message != "job0" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if _, err := json.Marshal(r); err != nil {
			t.Error(err)
		}

		for _, j := range jobs {
			job, err := l.Find(j.ID)
			if err != nil {
				t.Error(err)
			}
			if job.Result.Message != j.Result.Message {
				t.Errorf("Wrong job: %v", job)
			}

			if err := l.Delete(j.ID); err != nil {
				t.Error(err)
			}
			if _, err := l.Find(j.ID); err != sql.ErrNoRows {
				t.Error("Deleted job should not be found")
			}
		}
	}()
}

func TestWorkerStats(t *testing.T) {
	if test.If("driver", "in-memory") { // not supported
		return
	}

	queueName := "service_worker_stats_test_queue"
	maxWorkers := uint(120)

	svc := newService()

	{
		q := &model.Queue{Name: queueName, MaxWorkers: maxWorkers}
		err := svc.AddJobQueue(q)
		if err != nil {
			t.Error(err)
		}
	}
	time.Sleep(500 * time.Millisecond) // wait for up

	q, ok := svc.GetJobQueue(queueName)
	if !ok {
		t.Error("Should return the defined queue")
	}

	{
		stats := q.WorkerStats()
		if stats.TotalWorkers != int64(maxWorkers) {
			t.Error("Should return defined value")
		}
	}

	svc.DeleteJobQueue(queueName)
	<-svc.Stop()

	{
		stats := q.WorkerStats()
		if stats.TotalWorkers != 0 {
			t.Error("Should return 0 for a stopped queue")
		}
	}
}

type incomingJob struct {
	category   string
	url        string
	payload    string
	nextDelay  uint64
	retryDelay uint
	retryCount uint
}

func (job *incomingJob) Category() string {
	return job.category
}

func (job *incomingJob) URL() string {
	return job.url
}

func (job *incomingJob) Payload() string {
	return job.payload
}

func (job *incomingJob) NextDelay() uint64 {
	return job.nextDelay
}

func (job *incomingJob) NextTry() uint64 {
	return uint64(0)
}

func (job *incomingJob) RetryCount() uint {
	return job.retryCount
}

func (job *incomingJob) RetryDelay() uint {
	return job.retryDelay
}

func (job *incomingJob) Timeout() uint {
	return uint(0)
}

func newService() *Service {
	return NewService(repository.NewRepositories())
}

type testServer struct {
	worker *testWorker
	server *httptest.Server
	t      *testing.T
}

func newTestWorker(t *testing.T) *testServer {
	w := &testWorker{request: make(chan string, 1000)}
	return &testServer{
		worker: w,
		server: httptest.NewServer(w),
		t:      t,
	}
}

func (s *testServer) url() string {
	return s.server.URL
}

func (s *testServer) wait(dur time.Duration) string {
	select {
	case payload := <-s.worker.request:
		return payload
	case <-time.After(dur):
		s.t.Error("Timeout")
		return ""
	}
}

func (s *testServer) close() {
	s.server.Close()
}

type testWorker struct {
	payload string
	request chan string
}

func (worker *testWorker) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)

	w.WriteHeader(200)
	var result jobqueue.Result
	if err := json.Unmarshal(buf.Bytes(), &result); err == nil {
		w.Write(buf.Bytes())
	}
	worker.request <- buf.String()
}
