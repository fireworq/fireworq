package factory

import (
	"testing"

	"github.com/fireworq/fireworq/model"
	"github.com/fireworq/fireworq/test"
)

func TestMain(m *testing.M) {
	test.RunAll(m)
}

func TestQueue(t *testing.T) {
	repo := NewRepositories()

	{
		qs, err := repo.Queue.FindAll()
		if err != nil {
			t.Error(err)
		}
		if len(qs) != 0 {
			t.Error("There should be no queue at first")
		}
	}

	if u, err := repo.Queue.Add(&model.Queue{Name: "repo_queue_test_queue_1"}); !u || err != nil {
		t.Errorf("updated = %v (should be true), error: %s", u, err)
	}
	if u, err := repo.Queue.Add(&model.Queue{Name: "repo_queue_test_queue_2", MaxWorkers: 1000}); !u || err != nil {
		t.Errorf("updated = %v (should be true), error: %s", u, err)
	}
	if u, err := repo.Queue.Add(&model.Queue{
		Name:                   "repo_queue_test_queue_3",
		PollingInterval:        300,
		MaxWorkers:             10,
		MaxDispatchesPerSecond: 2.5,
		MaxBurstSize:           5,
	}); !u || err != nil {
		t.Errorf("updated = %v (should be true), error: %s", u, err)
	}

	{
		qs, err := repo.Queue.FindAll()
		if err != nil {
			t.Error(err)
		}
		if len(qs) != 3 {
			t.Error("There should be defined queues")
		}

		if qs[0].Name != "repo_queue_test_queue_1" {
			t.Error("Defined queues can be retrieved in name order")
		}
		if q := qs[0]; q.PollingInterval != 0 || q.MaxWorkers != 0 ||
			q.MaxDispatchesPerSecond != 0.0 || q.MaxBurstSize != 0 {
			t.Errorf("Defined queues can be retrieved: %#v", q)
		}

		if qs[1].Name != "repo_queue_test_queue_2" {
			t.Error("Defined queues can be retrieved in name order")
		}
		if q := qs[1]; q.PollingInterval != 0 || q.MaxWorkers != 1000 ||
			q.MaxDispatchesPerSecond != 0.0 || q.MaxBurstSize != 0 {
			t.Errorf("Defined queues can be retrieved: %#v", q)
		}

		if qs[2].Name != "repo_queue_test_queue_3" {
			t.Error("Defined queues can be retrieved in name order")
		}
		if q := qs[2]; q.PollingInterval != 300 || q.MaxWorkers != 10 ||
			q.MaxDispatchesPerSecond != 2.5 || q.MaxBurstSize != 5 {
			t.Errorf("Defined queues can be retrieved: %#v", q)
		}
	}

	{
		q, err := repo.Queue.FindByName("repo_queue_test_queue_2")
		if err != nil {
			t.Error(err)
		}
		if q.Name != "repo_queue_test_queue_2" {
			t.Error("Defined queue can be retrieved by name")
		}
		if q.PollingInterval != 0 || q.MaxWorkers != 1000 ||
			q.MaxDispatchesPerSecond != 0.0 || q.MaxBurstSize != 0 {
			t.Errorf("Defined queues can be retrieved by name: %#v", q)
		}
	}

	{
		q, err := repo.Queue.FindByName("repo_queue_test_queue_3")
		if err != nil {
			t.Error(err)
		}
		if q.Name != "repo_queue_test_queue_3" {
			t.Error("Defined queue can be retrieved by name")
		}
		if q.PollingInterval != 300 || q.MaxWorkers != 10 ||
			q.MaxDispatchesPerSecond != 2.5 || q.MaxBurstSize != 5 {
			t.Errorf("Defined queues can be retrieved by name: %#v", q)
		}
	}

	revision, err := repo.Queue.Revision()
	if err != nil {
		t.Error(err)
	}

	if u, err := repo.Queue.Add(&model.Queue{Name: "repo_queue_test_queue_1"}); u || err != nil {
		t.Errorf("updated = %v (should be false), error: %s", u, err)
	}
	if u, err := repo.Queue.Add(&model.Queue{Name: "repo_queue_test_queue_2", MaxWorkers: 1000}); u || err != nil {
		t.Errorf("updated = %v (should be false), error: %s", u, err)
	}

	revision1, err := repo.Queue.Revision()
	if err != nil {
		t.Error(err)
	}
	if revision1 != revision {
		t.Errorf("Revision %d != %d", revision1, revision)
	}

	if u, err := repo.Queue.Add(&model.Queue{Name: "repo_queue_test_queue_2", MaxWorkers: 100}); !u || err != nil {
		t.Errorf("updated = %v (should be true), error: %s", u, err)
	}

	revision2, err := repo.Queue.Revision()
	if err != nil {
		t.Error(err)
	}
	if revision2 <= revision {
		t.Errorf("Revision !(%d > %d)", revision2, revision)
	}

	if err := repo.Queue.DeleteByName("repo_queue_test_queue_1"); err != nil {
		t.Error(err)
	}

	{
		q, err := repo.Queue.FindByName("repo_queue_test_queue_1")
		if err == nil {
			t.Error("Deleted queue should not be found")
		}
		if q != nil {
			t.Error("Deleted queue should not be found")
		}
	}

	{
		qs, err := repo.Queue.FindAll()
		if err != nil {
			t.Error(err)
		}
		if len(qs) != 2 {
			t.Error("There should be defined queues")
		}

		if qs[0].Name != "repo_queue_test_queue_2" {
			t.Error("Defined queues can be retrieved in name order")
		}
		if qs[1].Name != "repo_queue_test_queue_3" {
			t.Error("Defined queues can be retrieved in name order")
		}
	}

	if err := repo.Queue.DeleteByName("repo_queue_test_queue_1"); err != nil {
		t.Errorf("Delete queue again should not fail: %s", err)
	}

	if err := repo.Queue.DeleteByName("repo_queue_test_queue_2"); err != nil {
		t.Error(err)
	}

	if err := repo.Queue.DeleteByName("repo_queue_test_queue_3"); err != nil {
		t.Error(err)
	}

	{
		qs, err := repo.Queue.FindAll()
		if err != nil {
			t.Error(err)
		}
		if len(qs) != 0 {
			t.Error("There should be no queues")
		}
	}
}

func TestRouting(t *testing.T) {
	repo := NewRepositories()

	{
		rs, err := repo.Routing.FindAll()
		if err != nil {
			t.Error(err)
		}
		if len(rs) != 0 {
			t.Error("There should be no routing at first")
		}
	}

	if _, err := repo.Queue.Add(&model.Queue{Name: "repo_routing_test_queue_1"}); err != nil {
		t.Error(err)
	}
	if _, err := repo.Queue.Add(&model.Queue{Name: "repo_routing_test_queue_2"}); err != nil {
		t.Error(err)
	}

	if err := repo.Routing.Add("repo_routing_test_A", "repo_routing_test_queue_1"); err != nil {
		t.Error(err)
	}
	if err := repo.Routing.Add("repo_routing_test_B", "repo_routing_test_queue_1"); err != nil {
		t.Error(err)
	}
	if err := repo.Routing.Add("repo_routing_test_C", "repo_routing_test_queue_2"); err != nil {
		t.Error(err)
	}

	{
		rs, err := repo.Routing.FindAll()
		if err != nil {
			t.Error(err)
		}
		if len(rs) != 3 {
			t.Error("There should be defined routings")
		}
	}

	{
		q := repo.Routing.FindQueueNameByJobCategory("repo_routing_test_A")
		if q != "repo_routing_test_queue_1" {
			t.Errorf("Wrong queue: %s", q)
		}
	}

	if err := repo.Routing.DeleteByJobCategory("repo_routing_test_B"); err != nil {
		t.Error(t)
	}

	{
		q := repo.Routing.FindQueueNameByJobCategory("repo_routing_test_B")
		if q != "" {
			t.Error("Deleted routing should have no queue")
		}
	}

	if err := repo.Routing.Reload(); err != nil {
		t.Error(err)
	}

	_, err := repo.Routing.Revision()
	if err != nil {
		t.Error(err)
	}
}
