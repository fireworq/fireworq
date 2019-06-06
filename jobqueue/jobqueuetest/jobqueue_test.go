package jobqueuetest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/jobqueue/factory"
	"github.com/fireworq/fireworq/model"
	"github.com/fireworq/fireworq/test"
)

func TestMain(m *testing.M) {
	test.RunAll(m)
}

func TestRecovering(t *testing.T) {
	if test.If("driver", "in-memory") { // not supported
		return
	}

	name := "jobqueue_recoveringtest"

	jq1 := start(&model.Queue{Name: name, MaxWorkers: 10})

	jq2 := start(&model.Queue{Name: name, MaxWorkers: 10})

	if !jq1.IsActive() {
		t.Error("The first job queue should be active")
	}

	jobs := make([]incomingJob, 10)
	for i, j := range jobs {
		j.url = fmt.Sprintf("job%d", i)
		jq1.Push(&j)
		time.Sleep(10 * time.Millisecond)
	}
	jq1.Pop(4)

	// jobs are not completed

	<-jq1.Stop()
	defer func() { <-jq2.Stop() }()

	ch := make(chan struct{})
	go func() {
	Loop:
		for {
			if jq2.IsActive() {
				break Loop
			}
			time.Sleep(10 * time.Millisecond)
		}
		ch <- struct{}{}
	}()

	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Error("A backup jobqueue should be activated")
	}
	time.Sleep(100 * time.Millisecond)

	launched, err := jq2.Pop(1)
	if err != nil {
		t.Error(err)
	}
	jq2.Complete(launched[0], &jobqueue.Result{
		Status: jobqueue.ResultStatusSuccess,
	})

	func() {
		ins, ok := jq2.Inspector()
		if !ok {
			t.Error("Cannot get the inspector")
		}
		r, err := ins.FindAllGrabbed(10, "")
		if err != nil {
			t.Error(err)
		}
		jobs := r.Jobs
		if len(jobs) != 0 {
			t.Errorf("%d jobs should have been recovered", len(jobs))
		}
	}()
}

func TestInspecting(t *testing.T) {
	queueName := "jobqueue_inspecting_test_queue"

	jq := start(&model.Queue{Name: queueName, MaxWorkers: 10})
	defer func() { <-jq.Stop() }()

	jobs := make([]incomingJob, 10)
	jobs[2].nextDelay = 500000
	jobs[5].nextDelay = 600000
	jobs[8].nextDelay = 500000
	jobs[6].payload = "$foo"
	for i, j := range jobs {
		j.url = fmt.Sprintf("job%d", i)
		jq.Push(&j)
		time.Sleep(10 * time.Millisecond)
	}

	jq.Pop(4)

	ins, ok := jq.Inspector()
	if test.If("driver", "in-memory") {
		if ins != nil || ok {
			t.Error("Implemented?")
		}
		return
	}
	if !ok {
		t.Error("Cannot get the inspector")
	}

	func() {
		r, err := ins.FindAllWaiting(2, "")
		if err != nil {
			t.Error(err)
		}
		jobs := r.Jobs
		if len(jobs) != 2 {
			t.Errorf("The number of jobs is incorrect: %d", len(jobs))
		}
		if jobs[0].URL != "job9" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if jobs[1].URL != "job7" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}

		r, err = ins.FindAllWaiting(2, r.NextCursor)
		if err != nil {
			t.Error(err)
		}
		if r.NextCursor != "" {
			t.Error("Cursor should be empty when there is no more jobs")
		}
		jobs = append(jobs, r.Jobs...)
		if len(jobs) != 3 {
			t.Errorf("The number of jobs is incorrect: %d", len(jobs))
		}
		if jobs[2].URL != "job6" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if _, err := json.Marshal(r); err != nil {
			t.Error(err)
		}

		for _, j := range jobs {
			job, err := ins.Find(j.ID)
			if err != nil {
				t.Error(err)
			}
			if job.URL != j.URL {
				t.Errorf("Wrong job: %v", job)
			}
		}
	}()

	func() {
		r, err := ins.FindAllGrabbed(3, "")
		if err != nil {
			t.Error(err)
		}
		jobs := r.Jobs
		if len(jobs) != 3 {
			t.Errorf("The number of jobs is incorrect: %d", len(jobs))
		}
		if jobs[0].URL != "job4" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if jobs[1].URL != "job3" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if jobs[2].URL != "job1" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}

		r, err = ins.FindAllGrabbed(3, r.NextCursor)
		if err != nil {
			t.Error(err)
		}
		if r.NextCursor != "" {
			t.Error("Cursor should be empty when there is no more jobs")
		}
		jobs = append(jobs, r.Jobs...)
		if len(jobs) != 4 {
			t.Errorf("The number of jobs is incorrect: %d", len(jobs))
		}
		if jobs[3].URL != "job0" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if _, err := json.Marshal(r); err != nil {
			t.Error(err)
		}

		for _, j := range jobs {
			job, err := ins.Find(j.ID)
			if err != nil {
				t.Error(err)
			}
			if job.URL != j.URL {
				t.Errorf("Wrong job: %v", job)
			}
		}
	}()

	func() {
		r, err := ins.FindAllDeferred(2, "")
		if err != nil {
			t.Error(err)
		}
		jobs := r.Jobs
		if len(jobs) != 2 {
			t.Errorf("The number of jobs is incorrect: %d", len(jobs))
		}
		if jobs[0].URL != "job5" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if jobs[1].URL != "job8" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}

		r, err = ins.FindAllDeferred(2, r.NextCursor)
		if err != nil {
			t.Error(err)
		}
		if r.NextCursor != "" {
			t.Error("Cursor should be empty when there is no more jobs")
		}
		jobs = append(jobs, r.Jobs...)
		if len(jobs) != 3 {
			t.Errorf("The number of jobs is incorrect: %d", len(jobs))
		}
		if jobs[2].URL != "job2" {
			t.Errorf("Invalid order of jobs: %v", jobs)
		}
		if _, err := json.Marshal(r); err != nil {
			t.Error(err)
		}

		for _, j := range jobs {
			job, err := ins.Find(j.ID)
			if err != nil {
				t.Error(err)
			}
			if job.URL != j.URL {
				t.Errorf("Wrong job: %v", job)
			}
		}
	}()

	func() {
		r, err := ins.FindAllWaiting(10, "")
		if err != nil {
			t.Error(err)
		}
		jobs := r.Jobs
		if len(jobs) != 3 {
			t.Errorf("The number of jobs is incorrect: %d", len(jobs))
		}

		for _, j := range jobs {
			if err := ins.Delete(j.ID); err != nil {
				t.Error(err)
			}
			if _, err := ins.Find(j.ID); err != sql.ErrNoRows {
				t.Error("Deleted job should not be found")
			}
		}
	}()

	func() {
		r, err := ins.FindAllGrabbed(10, "")
		if err != nil {
			t.Error(err)
		}
		jobs := r.Jobs
		if len(jobs) != 4 {
			t.Errorf("The number of jobs is incorrect: %d", len(jobs))
		}

		for _, j := range jobs {
			if err := ins.Delete(j.ID); err != nil {
				t.Error(err)
			}
			if _, err := ins.Find(j.ID); err != sql.ErrNoRows {
				t.Error("Deleted job should not be found")
			}
		}
	}()

	func() {
		r, err := ins.FindAllDeferred(10, "")
		if err != nil {
			t.Error(err)
		}
		jobs := r.Jobs
		if len(jobs) != 3 {
			t.Errorf("The number of jobs is incorrect: %d", len(jobs))
		}

		for _, j := range jobs {
			if err := ins.Delete(j.ID); err != nil {
				t.Error(err)
			}
			if _, err := ins.Find(j.ID); err != sql.ErrNoRows {
				t.Error("Deleted job should not be found")
			}
		}
	}()
}

func TestNodeInfo(t *testing.T) {
	queueName := "jobqueue_node_info_test_queue"

	jq := start(&model.Queue{Name: queueName, MaxWorkers: 10})
	defer func() { <-jq.Stop() }()

	node, err := jq.Node()
	if test.If("driver", "in-memory") {
		if node != nil {
			t.Error("Implemented?")
		}
		return
	}
	if err != nil {
		t.Error(err)
	}
	if node == nil {
		t.Error("There should be a node for a queue")
	}
	if len(node.Host) <= 0 {
		t.Error("There should be a host name for a node")
	}
}

func TestStats(t *testing.T) {
	queueName := "jobqueue_stats_test_queue"

	jq := start(&model.Queue{Name: queueName, MaxWorkers: 10})
	defer func() { <-jq.Stop() }()
	time.Sleep(100 * time.Millisecond) // wait for up

	qStats := jq.Stats()
	if qStats.TotalPushes != 0 || qStats.TotalPops != 0 || qStats.TotalCompletes != 0 || qStats.TotalFailures != 0 || qStats.TotalPermanentFailures != 0 || qStats.PushesPerSecond != 0 || qStats.PopsPerSecond != 0 {
		t.Error("Stats values must be zero before doing nothing")
	}

	jobs := make([]incomingJob, 10)
	jobs[2].nextDelay = 500000
	jobs[5].nextDelay = 600000
	jobs[8].nextDelay = 500000
	for i, j := range jobs {
		j.url = fmt.Sprintf("job%d", i)
		jq.Push(&j)
		time.Sleep(10 * time.Millisecond)
	}

	launched, err := jq.Pop(6)
	if err != nil {
		t.Error(err)
	}

	qStats = jq.Stats()
	if qStats.TotalPushes != 10 {
		t.Error("Stats should report the number of pushed jobs")
	}
	if qStats.TotalPops != 6 {
		t.Error("Stats should report the number of poped jobs")
	}
	if qStats.TotalCompletes != 0 || qStats.TotalFailures != 0 || qStats.TotalPermanentFailures != 0 {
		t.Error("Stats values must be zero before doing nothing")
	}

	jq.Complete(launched[0], &jobqueue.Result{
		Status: jobqueue.ResultStatusSuccess,
	})
	jq.Complete(launched[1], &jobqueue.Result{
		Status: jobqueue.ResultStatusPermanentFailure,
	})
	jq.Complete(launched[2], &jobqueue.Result{
		Status: jobqueue.ResultStatusSuccess,
	})
	jq.Complete(launched[3], &jobqueue.Result{
		Status: jobqueue.ResultStatusFailure,
	})
	jq.Complete(launched[4], &jobqueue.Result{
		Status: jobqueue.ResultStatusSuccess,
	})

	qStats = jq.Stats()
	if qStats.TotalPushes != 10 {
		t.Error("Stats should report the number of pushed jobs")
	}
	if qStats.TotalPops != 6 {
		t.Error("Stats should report the number of poped jobs")
	}
	if qStats.TotalCompletes != 5 {
		t.Error("Stats should report the number of completed jobs")
	}
	if qStats.TotalFailures != 2 {
		t.Error("Stats should report the number of failed jobs")
	}
	if qStats.TotalPermanentFailures != 1 {
		t.Error("Stats should report the number of permanently failed jobs")
	}
}

func start(q *model.Queue) jobqueue.JobQueue {
	impl := factory.NewImpl(q)
	jq := jobqueue.Start(q, impl)
	time.Sleep(500 * time.Millisecond) // wait for up
	return jq
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
