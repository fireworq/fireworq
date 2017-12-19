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

	if err := repo.Queue.Add(&model.Queue{Name: "repo_queue_test_queue_1"}); err != nil {
		t.Error(err)
	}
	if err := repo.Queue.Add(&model.Queue{Name: "repo_queue_test_queue_2", MaxWorkers: 1000}); err != nil {
		t.Error(err)
	}

	{
		qs, err := repo.Queue.FindAll()
		if err != nil {
			t.Error(err)
		}
		if len(qs) != 2 {
			t.Error("There should be defined queues")
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
		if q.MaxWorkers != 1000 {
			t.Error("Defined queue can be retrieved by name")
		}
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

	_, err := repo.Queue.Revision()
	if err != nil {
		t.Error(err)
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

	if err := repo.Queue.Add(&model.Queue{Name: "repo_routing_test_queue_1"}); err != nil {
		t.Error(err)
	}
	if err := repo.Queue.Add(&model.Queue{Name: "repo_routing_test_queue_2"}); err != nil {
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
