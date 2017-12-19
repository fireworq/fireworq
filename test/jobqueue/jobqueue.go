package jqtest

import (
	"testing"
	"time"

	"github.com/fireworq/fireworq/jobqueue"
)

type job struct {
	category   string
	url        string
	payload    string
	retryCount uint
	retryDelay uint
	timeout    uint
}

func (j *job) Category() string {
	return j.category
}

func (j *job) URL() string {
	return j.url
}

func (j *job) Payload() string {
	return j.payload
}

func (j *job) NextDelay() uint64 {
	return 1
}

func (j *job) NextTry() uint64 {
	return uint64(time.Now().UnixNano()/int64(time.Millisecond)) + j.NextDelay()
}

func (j *job) RetryCount() uint {
	return j.retryCount
}

func (j *job) RetryDelay() uint {
	return j.retryDelay
}

func (j *job) Timeout() uint {
	return j.timeout
}

const retryCount = 3

func newTestJob(category, url, data string) jobqueue.IncomingJob {
	time.Sleep(10 * time.Millisecond)
	return &job{
		category:   category,
		url:        url,
		payload:    data,
		retryCount: retryCount,
		retryDelay: 0,
		timeout:    1,
	}
}

type nextJob struct {
	jobqueue.Job
	nextDelay uint64
}

func (j *nextJob) NextDelay() uint64 {
	return j.nextDelay
}

func (j *nextJob) NextTry() uint64 {
	return uint64(time.Now().UnixNano()/int64(time.Millisecond)) + j.NextDelay()
}

func (j *nextJob) RetryCount() uint {
	return j.Job.RetryCount() - 1
}

func (j *nextJob) FailCount() uint {
	return j.Job.FailCount() + 1
}

// Subtest is an interface of a test function where the queue is
// assumed to be empty before running the test.
type Subtest func(t *testing.T, jq jobqueue.Impl)

// SubtestRunner is an interface of a function that runs multiple
// subtests.  An instance of this interface is required to be a
// function which runs specified subtests with truncating data store
// contents before running each subtest.
type SubtestRunner func(t *testing.T, db, q string, tests []Subtest)

// TestSubtests runs predefined subtests by the specified runner.
func TestSubtests(t *testing.T, runner SubtestRunner) {
	runner(t, "fireworq_test", "test_queue", []Subtest{
		subtestActive,
		subtestEmpty,
		subtestPush1,
		subtestPop1,
		subtestPopOrder,
		subtestPopPartially,
		subtestPopMulti,
		subtestDelete1,
		subtestDeletePartially,
		subtestDeleteMulti,
		subtestUpdate1,
		subtestUpdatePartially,
		subtestUpdateMulti,
		subtestAsyncPop1,
		subtestAsyncDelete1,
		subtestAsyncUpdate1,
	})
}

func subtestActive(t *testing.T, jq jobqueue.Impl) {
	if !jq.IsActive() {
		t.Error("Queue must be active")
	}
}

func subtestEmpty(t *testing.T, jq jobqueue.Impl) {
	jobs, err := jq.Pop(1)
	if err != nil {
		t.Error(err)
	}
	if len(jobs) != 0 {
		t.Error("Queue must be empty at the beginning")
	}
}

func subtestPush1(t *testing.T, jq jobqueue.Impl) {
	j, err := jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	if err != nil {
		t.Errorf("Failed to push job: %s", err)
	}
	if j.Payload() != "1" {
		t.Errorf("Wrong job returned: %v", j)
	}
}

func subtestPop1(t *testing.T, jq jobqueue.Impl) {
	jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	time.Sleep(10 * time.Millisecond)

	jobs, err := jq.Pop(10)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 1 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}
	if jobs[0].Payload() != "1" {
		t.Errorf("Wrong job returned: %v", jobs[0])
	}
}

func subtestPopOrder(t *testing.T, jq jobqueue.Impl) {
	jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "2"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "3"))
	jq.Push(newTestJob("foo", "http://localhost/worker", "4"))
	time.Sleep(10 * time.Millisecond)

	jobs, err := jq.Pop(10)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 4 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}
	for i, num := range []string{"1", "2", "3", "4"} {
		if jobs[i].Payload() != num {
			t.Errorf("Wrong job returned: %v", jobs[i])
		}
	}
}

func subtestPopPartially(t *testing.T, jq jobqueue.Impl) {
	jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "2"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "3"))
	jq.Push(newTestJob("foo", "http://localhost/worker", "4"))
	time.Sleep(10 * time.Millisecond)

	jobs, err := jq.Pop(3)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 3 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}
	for i, num := range []string{"1", "2", "3"} {
		if jobs[i].Payload() != num {
			t.Errorf("Wrong job returned: %v", jobs[i])
		}
	}
}

func subtestPopMulti(t *testing.T, jq jobqueue.Impl) {
	jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "2"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "3"))
	jq.Push(newTestJob("foo", "http://localhost/worker", "4"))
	jq.Push(newTestJob("foo", "http://localhost/worker", "5"))
	time.Sleep(10 * time.Millisecond)

	jobs, err := jq.Pop(3)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 3 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}
	for i, num := range []string{"1", "2", "3"} {
		if jobs[i].Payload() != num {
			t.Errorf("Wrong job returned: %v", jobs[i])
		}
	}

	jobs, err = jq.Pop(3)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 2 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}
	for i, num := range []string{"4", "5"} {
		if jobs[i].Payload() != num {
			t.Errorf("Wrong job returned: %v", jobs[i])
		}
	}
}

func subtestDelete1(t *testing.T, jq jobqueue.Impl) {
	jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	time.Sleep(10 * time.Millisecond)

	jobs, err := jq.Pop(10)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 1 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}
	jq.Delete(jobs[0])

	if hasInspector, ok := jq.(jobqueue.HasInspector); ok {
		i := hasInspector.Inspector()

		r1, err := i.FindAllGrabbed(uint(100), "")
		if err != nil {
			t.Error(err)
		}
		if len(r1.Jobs) != 0 {
			t.Error("There must be no grabbed job in the queue")
		}

		r2, err := i.FindAllWaiting(uint(100), "")
		if len(r2.Jobs) != 0 {
			t.Error("There must be no waiting job in the queue")
		}

		r3, err := i.FindAllDeferred(uint(100), "")
		if len(r3.Jobs) != 0 {
			t.Error("There must be no deferred job in the queue")
		}
	}
}

func subtestDeletePartially(t *testing.T, jq jobqueue.Impl) {
	jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "2"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "3"))
	jq.Push(newTestJob("foo", "http://localhost/worker", "4"))
	time.Sleep(10 * time.Millisecond)

	jobs, err := jq.Pop(3)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 3 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}

	jq.Delete(jobs[0])
	jq.Delete(jobs[1])

	if hasInspector, ok := jq.(jobqueue.HasInspector); ok {
		i := hasInspector.Inspector()

		r1, err := i.FindAllGrabbed(uint(100), "")
		if err != nil {
			t.Error(err)
		}
		if len(r1.Jobs) != 1 {
			t.Error("There must be only one grabbed job in the queue")
		}

		r2, err := i.FindAllWaiting(uint(100), "")
		if len(r2.Jobs) != 1 {
			t.Error("There must be one waiting job in the queue")
		}

		r3, err := i.FindAllDeferred(uint(100), "")
		if len(r3.Jobs) != 0 {
			t.Error("There must be no deferred job in the queue")
		}
	}
}

func subtestDeleteMulti(t *testing.T, jq jobqueue.Impl) {
	jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "2"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "3"))
	jq.Push(newTestJob("foo", "http://localhost/worker", "4"))
	jq.Push(newTestJob("foo", "http://localhost/worker", "5"))
	time.Sleep(10 * time.Millisecond)

	jobs, err := jq.Pop(3)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 3 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}

	jq.Delete(jobs[0])
	jq.Delete(jobs[1])

	if hasInspector, ok := jq.(jobqueue.HasInspector); ok {
		i := hasInspector.Inspector()

		r1, err := i.FindAllGrabbed(uint(100), "")
		if err != nil {
			t.Error(err)
		}
		if len(r1.Jobs) != 1 {
			t.Error("There must be only one grabbed job in the queue")
		}

		r2, err := i.FindAllWaiting(uint(100), "")
		if len(r2.Jobs) != 2 {
			t.Error("There must be two waiting jobs in the queue")
		}

		r3, err := i.FindAllDeferred(uint(100), "")
		if len(r3.Jobs) != 0 {
			t.Error("There must be no deferred jobs in the queue")
		}
	}

	jobs, err = jq.Pop(3)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 2 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}
	for _, j := range jobs {
		jq.Delete(j)
	}

	if hasInspector, ok := jq.(jobqueue.HasInspector); ok {
		i := hasInspector.Inspector()

		r1, err := i.FindAllGrabbed(uint(100), "")
		if err != nil {
			t.Error(err)
		}
		if len(r1.Jobs) != 1 {
			t.Error("There must be only one grabbed job in the queue")
		}

		r2, err := i.FindAllWaiting(uint(100), "")
		if len(r2.Jobs) != 0 {
			t.Error("There must be no waiting job in the queue")
		}

		r3, err := i.FindAllDeferred(uint(100), "")
		if len(r3.Jobs) != 0 {
			t.Error("There must be no deferred job in the queue")
		}
	}
}

func subtestUpdate1(t *testing.T, jq jobqueue.Impl) {
	jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	time.Sleep(10 * time.Millisecond)

	jobs, err := jq.Pop(10)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 1 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}
	count := jobs[0].RetryCount()
	failed := jobs[0].FailCount()
	jq.Update(jobs[0], &nextJob{jobs[0], 0})
	time.Sleep(10 * time.Millisecond)

	jobs, err = jq.Pop(10)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 1 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}
	if jobs[0].Payload() != "1" {
		t.Errorf("Wrong job returned: %v", jobs[0])
	}
	if jobs[0].RetryCount() != count-1 {
		t.Errorf("Invalid retry count: %d", jobs[0].RetryCount())
	}
	if jobs[0].FailCount() != failed+1 {
		t.Errorf("Invalid fail count: %d", jobs[0].FailCount())
	}
}

func subtestUpdatePartially(t *testing.T, jq jobqueue.Impl) {
	jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "2"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "3"))
	jq.Push(newTestJob("foo", "http://localhost/worker", "4"))
	time.Sleep(10 * time.Millisecond)

	jobs, err := jq.Pop(3)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 3 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}

	count := []uint{jobs[0].RetryCount(), jobs[1].RetryCount()}
	failed := []uint{jobs[0].FailCount(), jobs[1].FailCount()}

	jq.Update(jobs[0], &nextJob{jobs[0], 1})
	jq.Update(jobs[1], &nextJob{jobs[1], 2})
	time.Sleep(10 * time.Millisecond)

	jobs, err = jq.Pop(10)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 3 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}
	for i, j := range []string{"1", "2"} {
		if jobs[i+1].Payload() != j {
			t.Errorf("Wrong job returned: %s", jobs[i+1].Payload())
		}
		if jobs[i+1].RetryCount() != count[i]-1 {
			t.Errorf("Invalid retry count: %d", jobs[i+1].RetryCount())
		}
		if jobs[i+1].FailCount() != failed[i]+1 {
			t.Errorf("Invalid fail count: %d", jobs[i+1].FailCount())
		}
	}

	if jobs[0].Payload() != "4" {
		t.Errorf("Wrong job returned: %s", jobs[0].Payload())
	}
	if jobs[0].RetryCount() != retryCount {
		t.Errorf("Invalid retry count: %d", jobs[0].RetryCount())
	}
	if jobs[0].FailCount() != 0 {
		t.Errorf("Invalid fail count: %d", jobs[0].FailCount())
	}
}

func subtestUpdateMulti(t *testing.T, jq jobqueue.Impl) {
	jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "2"))
	jq.Push(newTestJob("bar", "http://localhost/worker", "3"))
	jq.Push(newTestJob("foo", "http://localhost/worker", "4"))
	jq.Push(newTestJob("foo", "http://localhost/worker", "5"))
	time.Sleep(10 * time.Millisecond)

	jobs, err := jq.Pop(3)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 3 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}

	jq.Update(jobs[0], &nextJob{jobs[0], 1})
	jq.Update(jobs[1], &nextJob{jobs[1], 2})
	time.Sleep(10 * time.Millisecond)

	jobs, err = jq.Pop(10)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 4 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}
	jq.Update(jobs[2], &nextJob{jobs[2], uint64(1)})
	jq.Update(jobs[1], &nextJob{jobs[1], uint64(2)})
	jq.Update(jobs[3], &nextJob{jobs[3], uint64(3)})
	jq.Update(jobs[0], &nextJob{jobs[0], uint64(4)})
	time.Sleep(10 * time.Millisecond)

	jobs, err = jq.Pop(10)
	if err != nil {
		t.Errorf("Failed to pop job: %s", err)
	}
	if len(jobs) != 4 {
		t.Errorf("Wrong queue length: %d", len(jobs))
	}
	for i, j := range []string{"1", "5", "2", "4"} {
		count := uint(1)
		if i%2 == 0 {
			count++
		}
		if jobs[i].Payload() != j {
			t.Errorf("Wrong job returned: %v", jobs[i])
		}
		if jobs[i].RetryCount() != retryCount-count {
			t.Errorf("Invalid retry count: %d", jobs[i].RetryCount())
		}
		if jobs[i].FailCount() != count {
			t.Errorf("Invalid fail count: %d", jobs[i].FailCount())
		}
	}
}

func subtestAsyncPop1(t *testing.T, jq jobqueue.Impl) {
	jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	time.Sleep(10 * time.Millisecond)

	done := make(chan struct{})

	go func() {
		jobs, err := jq.Pop(10)
		if err != nil {
			t.Errorf("Failed to pop job: %s", err)
		}
		if len(jobs) != 1 {
			t.Errorf("Wrong queue length: %d", len(jobs))
		}
		if jobs[0].Payload() != "1" {
			t.Errorf("Wrong job returned: %v", jobs[0])
		}
		done <- struct{}{}
	}()

	<-done
}

func subtestAsyncDelete1(t *testing.T, jq jobqueue.Impl) {
	jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	time.Sleep(10 * time.Millisecond)

	done := make(chan struct{})
	job := make(chan jobqueue.Job)

	go func() {
		jobs, err := jq.Pop(10)
		if err != nil {
			t.Errorf("Failed to pop job: %s", err)
		}
		if len(jobs) != 1 {
			t.Errorf("Wrong queue length: %d", len(jobs))
		}

		job <- jobs[0]
	}()

	j := <-job
	go func(j jobqueue.Job) {
		jq.Delete(j)
		done <- struct{}{}
	}(j)

	<-done

	if hasInspector, ok := jq.(jobqueue.HasInspector); ok {
		i := hasInspector.Inspector()

		r1, err := i.FindAllGrabbed(uint(100), "")
		if err != nil {
			t.Error(err)
		}
		if len(r1.Jobs) != 0 {
			t.Error("There must be no grabbed job in the queue")
		}

		r2, err := i.FindAllWaiting(uint(100), "")
		if len(r2.Jobs) != 0 {
			t.Error("There must be no waiting job in the queue")
		}

		r3, err := i.FindAllDeferred(uint(100), "")
		if len(r3.Jobs) != 0 {
			t.Error("There must be no deferred job in the queue")
		}
	}
}

func subtestAsyncUpdate1(t *testing.T, jq jobqueue.Impl) {
	jq.Push(newTestJob("foo", "http://localhost/worker", "1"))
	time.Sleep(10 * time.Millisecond)

	done := make(chan struct{})
	job := make(chan jobqueue.Job)

	go func() {
		jobs, err := jq.Pop(10)
		if err != nil {
			t.Errorf("Failed to pop job: %s", err)
		}
		if len(jobs) != 1 {
			t.Errorf("Wrong queue length: %d", len(jobs))
		}
		job <- jobs[0]
	}()

	j := <-job
	count := j.RetryCount()
	failed := j.FailCount()
	go func(j jobqueue.Job) {
		jq.Update(j, &nextJob{j, 1})
		done <- struct{}{}
	}(j)

	<-done
	time.Sleep(10 * time.Millisecond)

	go func() {
		jobs, err := jq.Pop(10)
		if err != nil {
			t.Errorf("Failed to pop job: %s", err)
		}
		if len(jobs) != 1 {
			t.Errorf("Wrong queue length: %d", len(jobs))
		}
		if jobs[0].Payload() != "1" {
			t.Errorf("Wrong job returned: %v", jobs[0])
		}
		if jobs[0].RetryCount() != count-1 {
			t.Errorf("Invalid retry count: %d", jobs[0].RetryCount())
		}
		if jobs[0].FailCount() != failed+1 {
			t.Errorf("Invalid fail count: %d", jobs[0].FailCount())
		}
		done <- struct{}{}
	}()

	<-done
}
