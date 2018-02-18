package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/fireworq/fireworq/dispatcher"
	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/model"

	"github.com/golang/mock/gomock"
)

func TestGetQueueList(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.QueueRepository.EXPECT().
			FindAll().
			Return([]model.Queue{}, errors.New("FindAll() failure"))

		resp, err := http.Get(s.URL + "/queues")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("GET /queues should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.QueueRepository.EXPECT().
			FindAll().
			Return([]model.Queue{}, nil)

		resp, err := http.Get(s.URL + "/queues")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queues should succeed")
		}

		var qs []model.Queue
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &qs); err != nil {
			t.Error("GET /queues should return queue definitions")
		}
		if len(qs) != 0 {
			t.Error("GET /queues should return empty list if nothing defined")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		queues := []model.Queue{
			{
				Name:            "queue1",
				PollingInterval: 100,
				MaxWorkers:      10,
			},
			{
				Name:            "queue2",
				PollingInterval: 200,
				MaxWorkers:      20,
			},
			{
				Name:            "queue3",
				PollingInterval: 300,
				MaxWorkers:      30,
			},
		}
		mockApp.QueueRepository.EXPECT().
			FindAll().
			Return(queues, nil)

		resp, err := http.Get(s.URL + "/queues")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queues should succeed")
		}

		var qs []model.Queue
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &qs); err != nil {
			t.Error(err)
		}
		if len(qs) != len(queues) {
			t.Error("GET /queues should return defined queues")
		}
		for i, q := range qs {
			if q.Name != queues[i].Name || q.PollingInterval != queues[i].PollingInterval || q.MaxWorkers != queues[i].MaxWorkers {
				t.Error("GET /queues should return defined queues")
			}
		}
	}()
}

func TestGetQueueListStats(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.QueueRepository.EXPECT().
			FindAll().
			Return([]model.Queue{}, errors.New("FindAll() failure"))

		resp, err := http.Get(s.URL + "/queues/stats")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("GET /queues/stats should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.QueueRepository.EXPECT().
			FindAll().
			Return([]model.Queue{}, nil)

		resp, err := http.Get(s.URL + "/queues/stats")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queues/stats should succeed")
		}

		var m map[string]jobqueue.Stats
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &m); err != nil {
			t.Error("GET /queues/stats should return a map of stats")
		}
		if len(m) != 0 {
			t.Error("GET /queues/stats should return empty map if no queue is defined")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		queues := []model.Queue{
			{
				Name:            "queue1",
				PollingInterval: 100,
				MaxWorkers:      10,
			},
			{
				Name:            "queue2",
				PollingInterval: 200,
				MaxWorkers:      20,
			},
			{
				Name:            "queue3",
				PollingInterval: 300,
				MaxWorkers:      30,
			},
		}
		mockApp.QueueRepository.EXPECT().
			FindAll().
			Return(queues, nil)

		stats := map[string]*jobqueue.Stats{
			"queue1": {
				TotalPushes:     10,
				TotalPops:       8,
				TotalCompletes:  5,
				TotalFailures:   2,
				PushesPerSecond: 2,
				PopsPerSecond:   1,
			},
			"queue2": {
				TotalPushes:     100,
				TotalPops:       100,
				TotalCompletes:  32,
				TotalFailures:   0,
				PushesPerSecond: 10,
				PopsPerSecond:   10,
			},
			"queue3": {
				TotalPushes:     1,
				TotalPops:       0,
				TotalCompletes:  0,
				TotalFailures:   0,
				PushesPerSecond: 0,
				PopsPerSecond:   0,
			},
		}
		lchrStats := map[string]*dispatcher.Stats{
			"queue1": {
				TotalWorkers: 10,
				IdleWorkers:  10,
			},
			"queue2": {
				TotalWorkers: 20,
				IdleWorkers:  15,
			},
			"queue3": {
				TotalWorkers: 30,
				IdleWorkers:  20,
			},
		}

		mockJobQueue1 := NewMockJobQueue(ctrl)
		mockJobQueue1.EXPECT().
			Stats().
			Return(stats["queue1"])
		mockJobQueue1.EXPECT().
			IsActive().
			Return(true)

		mockApp.Service.EXPECT().
			GetJobQueue("queue1").
			Return(newMockRunningQueue(mockJobQueue1, lchrStats["queue1"]), true)

		mockJobQueue2 := NewMockJobQueue(ctrl)
		mockJobQueue2.EXPECT().
			Stats().
			Return(stats["queue2"])
		mockJobQueue2.EXPECT().
			IsActive().
			Return(false)

		mockApp.Service.EXPECT().
			GetJobQueue("queue2").
			Return(newMockRunningQueue(mockJobQueue2, lchrStats["queue2"]), true)

		mockJobQueue3 := NewMockJobQueue(ctrl)
		mockJobQueue3.EXPECT().
			Stats().
			Return(stats["queue3"])
		mockJobQueue3.EXPECT().
			IsActive().
			Return(false)

		mockApp.Service.EXPECT().
			GetJobQueue("queue3").
			Return(newMockRunningQueue(mockJobQueue3, lchrStats["queue3"]), true)

		resp, err := http.Get(s.URL + "/queues/stats")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queues/stats should succeed")
		}

		var m map[string]Stats
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &m); err != nil {
			t.Error("GET /queues/stats should return a map of stats")
		}
		if len(m) != len(queues) {
			t.Error("GET /queues/stats should return stats of defined queues")
		}
		for k, s := range m {
			if s.TotalPushes != stats[k].TotalPushes || s.TotalPops != stats[k].TotalPops || s.TotalCompletes != stats[k].TotalCompletes || s.TotalFailures != stats[k].TotalFailures || s.PushesPerSecond != stats[k].PushesPerSecond || s.PopsPerSecond != stats[k].PopsPerSecond {
				t.Error("GET /queues/stats should return stats of defined queues")
			}
		}
		if m["queue1"].ActiveNodes != 1 || m["queue2"].ActiveNodes != 0 || m["queue3"].ActiveNodes != 0 {
			t.Error("GET /queues/stats should return stats of defined queues")
		}
	}()
}

func TestGetQueue(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.QueueRepository.EXPECT().
			FindByName("test_queue1").
			Return(nil, errors.New("nothing found"))

		resp, err := http.Get(s.URL + "/queue/test_queue1")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("GET /queue/$name should return 404 for an undefinde queue")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		queue := &model.Queue{
			Name:       "test_queue2",
			MaxWorkers: 123,
		}
		mockApp.QueueRepository.EXPECT().
			FindByName(queue.Name).
			Return(queue, nil)

		resp, err := http.Get(s.URL + "/queue/" + queue.Name)
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name should succeed")
		}

		var q model.Queue
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &q); err != nil {
			t.Error(err)
		}
		if q.Name != queue.Name {
			t.Error("GET /queue/$name should return a defined queue")
		}
	}()
}

func TestPutQueue(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, _ := newMockServer(ctrl)
		defer s.Close()

		resp, err := putJSON(s.URL+"/queue/test_queue3", "foo")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Error("PUT /queue/$name should reject invalid input")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		queue := &model.Queue{
			Name:            "test_queue4",
			PollingInterval: 500,
			MaxWorkers:      50,
		}
		def := &model.Queue{
			PollingInterval: queue.PollingInterval,
			MaxWorkers:      queue.MaxWorkers,
		}

		mockApp.Service.EXPECT().
			AddJobQueue(gomock.Any()).
			Return(errors.New("AddJobQueue() failure"))

		resp, err := putJSON(s.URL+"/queue/test_queue", def)
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("PUT /queue/$name should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		queue := &model.Queue{
			Name:            "test_queue4",
			PollingInterval: 500,
			MaxWorkers:      50,
		}
		def := &model.Queue{
			PollingInterval: queue.PollingInterval,
			MaxWorkers:      queue.MaxWorkers,
		}

		mockApp.Service.EXPECT().
			AddJobQueue(gomock.Any()).
			Return(nil)

		resp, err := putJSON(s.URL+"/queue/"+queue.Name, def)
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("PUT /queue/$name should accept a queue definition")
		}

		var q model.Queue
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &q); err != nil {
			t.Error(err)
		}
		if q.Name != queue.Name || q.PollingInterval != queue.PollingInterval || q.MaxWorkers != queue.MaxWorkers {
			t.Error("PUT /queue/$name should return a defined queue")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		queue := &model.Queue{
			Name:            "test_queueX",
			PollingInterval: 500,
			MaxWorkers:      50,
		}

		mockApp.Service.EXPECT().
			AddJobQueue(gomock.Any()).
			Return(nil)

		resp, err := putJSON(s.URL+"/queue/"+"test_queue5", queue)
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("PUT /queue/$name should accept a queue definition")
		}

		var q model.Queue
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &q); err != nil {
			t.Error(err)
		}
		if q.Name != "test_queue5" || q.PollingInterval != queue.PollingInterval || q.MaxWorkers != queue.MaxWorkers {
			t.Error("PUT /queue/$name should return a defined queue")
		}
		if q.Name == queue.Name {
			t.Error("PUT /queue/$name should ignore the queue name in the body")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		queue := &model.Queue{Name: "test_queue5"}

		mockApp.Service.EXPECT().
			AddJobQueue(gomock.Any()).
			Return(nil)

		resp, err := putJSON(s.URL+"/queue/"+queue.Name, emptyObject{})
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("PUT /queue/$name should accept an empty definition")
		}

		var q model.Queue
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &q); err != nil {
			t.Error(err)
		}
		if q.Name != queue.Name {
			t.Error("PUT /queue/$name should return a defined queue")
		}
		if q.PollingInterval != 0 || q.MaxWorkers != 0 {
			t.Error("PUT /queue/$name should return a defined queue with default values")
		}
	}()
}

func TestDeleteQueue(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.QueueRepository.EXPECT().
			FindByName("test_queue1").
			Return(nil, errors.New("nothing found"))

		resp, err := httpDelete(s.URL + "/queue/test_queue1")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("DELETE /queue/$name should return 404 for an undefinde queue")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		queue := &model.Queue{
			Name:       "test_queue2",
			MaxWorkers: 123,
		}
		mockApp.QueueRepository.EXPECT().
			FindByName(queue.Name).
			Return(queue, nil)
		mockApp.Service.EXPECT().
			DeleteJobQueue(queue.Name).
			Return(errors.New("DeleteJobQueue() failure"))

		resp, err := httpDelete(s.URL + "/queue/" + queue.Name)
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("DELETE /queue/$name should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		queue := &model.Queue{
			Name:       "test_queue2",
			MaxWorkers: 123,
		}
		mockApp.QueueRepository.EXPECT().
			FindByName(queue.Name).
			Return(queue, nil)
		mockApp.Service.EXPECT().
			DeleteJobQueue(queue.Name).
			Return(nil)

		resp, err := httpDelete(s.URL + "/queue/" + queue.Name)
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("DELETE /queue/$name should succeed")
		}

		var q model.Queue
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &q); err != nil {
			t.Error(err)
		}
		if q.Name != queue.Name {
			t.Error("DELETE /queue/$name should return a deleted queue")
		}
	}()
}

func TestGetQueueNode(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(nil, false)

		resp, err := http.Get(s.URL + "/queue/queue1/node")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("GET /queue/$name/node should return 404 for an undefinde queue")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Node().
			Return(nil, nil)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/node")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("GET /queue/$name/node should return 404 if there is no node")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Node().
			Return(nil, errors.New("error"))

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/node")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("GET /queue/$name/node should return an error")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		node := &jobqueue.Node{
			ID:   "123",
			Host: "192.168.1.2",
		}

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Node().
			Return(node, nil)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/node")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name/node should succeed")
		}
		var result jobqueue.Node
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &result); err != nil {
			t.Error(err)
		}
		if result.ID != node.ID || result.Host != node.Host {
			t.Error("GET /queue/$name/node should return correct node info")
		}
	}()
}

func TestGetQueueStats(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.Service.EXPECT().
			GetJobQueue("queue1").
			Return(nil, false)

		resp, err := http.Get(s.URL + "/queue/queue1/stats")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("GET /queue/queue1/stats should return 404 if the is not found")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		stats := &jobqueue.Stats{
			TotalPushes:     10,
			TotalPops:       8,
			TotalCompletes:  5,
			TotalFailures:   2,
			PushesPerSecond: 2,
			PopsPerSecond:   1,
		}
		lchrStats := &dispatcher.Stats{
			TotalWorkers: 10,
			IdleWorkers:  10,
		}

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Stats().
			Return(stats)
		mockJobQueue.EXPECT().
			IsActive().
			Return(true)

		mockApp.Service.EXPECT().
			GetJobQueue("queue1").
			Return(newMockRunningQueue(mockJobQueue, lchrStats), true)

		resp, err := http.Get(s.URL + "/queue/queue1/stats")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/queue1/stats should succeed")
		}

		var result Stats
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &result); err != nil {
			t.Error("GET /queue/queue1/stats should return a stats")
		}
		if result.TotalPushes != stats.TotalPushes || result.TotalPops != stats.TotalPops || result.TotalCompletes != stats.TotalCompletes || result.TotalFailures != stats.TotalFailures || result.PushesPerSecond != stats.PushesPerSecond || result.PopsPerSecond != stats.PopsPerSecond || result.ActiveNodes != 1 {
			t.Error("GET /queue/queue1/stats should return stats of the queue")
		}
	}()
}

func TestGetQueueGrabbed(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(nil, false)

		resp, err := http.Get(s.URL + "/queue/queue1/grabbed")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("GET /queue/$name/grabbed should return 404 for an undefinde queue")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(nil, false)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/grabbed")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotImplemented {
			t.Error("GET /queue/$name/grabbed should return 501 if there is no failure log interface")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			FindAllGrabbed(gomock.Any(), "").
			Return(nil, errors.New("FindAllGrabbed() failure"))

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/grabbed")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("GET /queue/$name/grabbed should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		jobs := &jobqueue.InspectedJobs{
			Jobs: []jobqueue.InspectedJob{
				{
					ID:       3,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       2,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       1,
					Category: "test_job",
					URL:      "http://example.com/",
				},
			},
			NextCursor: "",
		}

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			FindAllGrabbed(gomock.Any(), "").
			Return(jobs, nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/grabbed")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name/grabbed should succeed")
		}

		var result jobqueue.InspectedJobs
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &result); err != nil {
			t.Error(err)
		}
		if len(result.Jobs) != len(jobs.Jobs) {
			t.Errorf("GET /queue/$name/grabbed should return grabbed jobs: %v", result)
		}
		for i, f := range result.Jobs {
			if f.ID != jobs.Jobs[i].ID {
				t.Errorf("GET /queue/$name/grabbed should return grabbed jobs: %v", f)
			}
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		jobs := &jobqueue.InspectedJobs{
			Jobs: []jobqueue.InspectedJob{
				{
					ID:       3,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       2,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       1,
					Category: "test_job",
					URL:      "http://example.com/",
				},
			},
			NextCursor: "",
		}

		limit := uint(123)
		cursor := "foo"

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			FindAllGrabbed(limit, cursor).
			Return(jobs, nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/grabbed?cursor=" + cursor + "&limit=" + strconv.Itoa(int(limit)))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name/grabbed should succeed")
		}

		var result jobqueue.InspectedJobs
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &result); err != nil {
			t.Error(err)
		}
		if len(result.Jobs) != len(jobs.Jobs) {
			t.Errorf("GET /queue/$name/grabbed should return grabbed jobs: %v", result)
		}
		for i, f := range result.Jobs {
			if f.ID != jobs.Jobs[i].ID {
				t.Errorf("GET /queue/$name/grabbed should return grabbed jobs: %v", f)
			}
		}
	}()
}

func TestGetQueueWaiting(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(nil, false)

		resp, err := http.Get(s.URL + "/queue/queue1/waiting")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("GET /queue/$name/waiting should return 404 for an undefinde queue")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(nil, false)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/waiting")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotImplemented {
			t.Error("GET /queue/$name/waiting should return 501 if there is no failure log interface")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			FindAllWaiting(gomock.Any(), "").
			Return(nil, errors.New("FindAllWating() failure"))

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/waiting")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("GET /queue/$name/waiting should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		jobs := &jobqueue.InspectedJobs{
			Jobs: []jobqueue.InspectedJob{
				{
					ID:       3,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       2,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       1,
					Category: "test_job",
					URL:      "http://example.com/",
				},
			},
			NextCursor: "",
		}

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			FindAllWaiting(gomock.Any(), "").
			Return(jobs, nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/waiting")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name/waiting should succeed")
		}

		var result jobqueue.InspectedJobs
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &result); err != nil {
			t.Error(err)
		}
		if len(result.Jobs) != len(jobs.Jobs) {
			t.Errorf("GET /queue/$name/waiting should return waiting jobs: %v", result)
		}
		for i, f := range result.Jobs {
			if f.ID != jobs.Jobs[i].ID {
				t.Errorf("GET /queue/$name/waiting should return waiting jobs: %v", f)
			}
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		jobs := &jobqueue.InspectedJobs{
			Jobs: []jobqueue.InspectedJob{
				{
					ID:       3,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       2,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       1,
					Category: "test_job",
					URL:      "http://example.com/",
				},
			},
			NextCursor: "",
		}

		limit := uint(123)
		cursor := "foo"

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			FindAllWaiting(limit, cursor).
			Return(jobs, nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/waiting?cursor=" + cursor + "&limit=" + strconv.Itoa(int(limit)))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name/waiting should succeed")
		}

		var result jobqueue.InspectedJobs
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &result); err != nil {
			t.Error(err)
		}
		if len(result.Jobs) != len(jobs.Jobs) {
			t.Errorf("GET /queue/$name/waiting should return waiting jobs: %v", result)
		}
		for i, f := range result.Jobs {
			if f.ID != jobs.Jobs[i].ID {
				t.Errorf("GET /queue/$name/waiting should return waiting jobs: %v", f)
			}
		}
	}()
}

func TestGetQueueDeferred(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(nil, false)

		resp, err := http.Get(s.URL + "/queue/queue1/deferred")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("GET /queue/$name/deferred should return 404 for an undefinde queue")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(nil, false)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/deferred")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotImplemented {
			t.Error("GET /queue/$name/deferred should return 501 if there is no failure log interface")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			FindAllDeferred(gomock.Any(), "").
			Return(nil, errors.New("FindAllDeferred() failure"))

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/deferred")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("GET /queue/$name/deferred should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		jobs := &jobqueue.InspectedJobs{
			Jobs: []jobqueue.InspectedJob{
				{
					ID:       3,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       2,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       1,
					Category: "test_job",
					URL:      "http://example.com/",
				},
			}, NextCursor: "",
		}

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			FindAllDeferred(gomock.Any(), "").
			Return(jobs, nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/deferred")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name/deferred should succeed")
		}

		var result jobqueue.InspectedJobs
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &result); err != nil {
			t.Error(err)
		}
		if len(result.Jobs) != len(jobs.Jobs) {
			t.Errorf("GET /queue/$name/deferred should return deferred jobs: %v", result)
		}
		for i, f := range result.Jobs {
			if f.ID != jobs.Jobs[i].ID {
				t.Errorf("GET /queue/$name/deferred should return deferred jobs: %v", f)
			}
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		jobs := &jobqueue.InspectedJobs{
			Jobs: []jobqueue.InspectedJob{
				{
					ID:       3,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       2,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       1,
					Category: "test_job",
					URL:      "http://example.com/",
				},
			},
			NextCursor: "",
		}

		limit := uint(123)
		cursor := "foo"

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			FindAllDeferred(limit, cursor).
			Return(jobs, nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/deferred?cursor=" + cursor + "&limit=" + strconv.Itoa(int(limit)))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name/deferred should succeed")
		}

		var result jobqueue.InspectedJobs
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &result); err != nil {
			t.Error(err)
		}
		if len(result.Jobs) != len(jobs.Jobs) {
			t.Errorf("GET /queue/$name/deferred should return deferred jobs: %v", result)
		}
		for i, f := range result.Jobs {
			if f.ID != jobs.Jobs[i].ID {
				t.Errorf("GET /queue/$name/deferred should return deferred jobs: %v", f)
			}
		}
	}()
}

func TestGetQueueJob(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, _ := newMockServer(ctrl)
		defer s.Close()

		resp, err := http.Get(s.URL + "/queue/queue1/job/a")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Error("GET /queue/$name/job/$id should reject invalid ID")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(nil, false)

		resp, err := http.Get(s.URL + "/queue/queue1/job/3")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("GET /queue/$name/job/$id should return 404 for an undefinde queue")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(nil, false)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/job/3")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotImplemented {
			t.Error("GET /queue/$name/job/$id should return 501 if there is no failure log interface")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			Find(uint64(3)).
			Return(nil, sql.ErrNoRows)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/job/3")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("GET /queue/$name/job/$id should 404 for an unknown job")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			Find(uint64(3)).
			Return(nil, errors.New("Find() failure"))

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/job/3")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("GET /queue/$name/job/$id should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		job := &jobqueue.InspectedJob{
			ID:       3,
			Category: "test_job",
			URL:      "http://example.com/",
		}

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			Find(uint64(3)).
			Return(job, nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/queue1/job/3")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name/job/$id should succeed")
		}

		var result jobqueue.InspectedJob
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &result); err != nil {
			t.Error(err)
		}
		if result.ID != job.ID {
			t.Errorf("GET /queue/$name/job/$id should return grabbed jobs: %v", result)
		}
	}()
}

func TestDeleteQueueJob(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, _ := newMockServer(ctrl)
		defer s.Close()

		resp, err := httpDelete(s.URL + "/queue/queue1/job/a")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Error("DELETE /queue/$name/job/$id should reject invalid ID")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(nil, false)

		resp, err := httpDelete(s.URL + "/queue/queue1/job/3")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("DELETE /queue/$name/job/$id should return 404 for an undefinde queue")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(nil, false)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := httpDelete(s.URL + "/queue/queue1/job/3")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotImplemented {
			t.Error("DELETE /queue/$name/job/$id should return 501 if there is no failure log interface")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			Find(uint64(3)).
			Return(nil, sql.ErrNoRows)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := httpDelete(s.URL + "/queue/queue1/job/3")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("DELETE /queue/$name/job/$id should return 404 for an unknown job")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			Find(uint64(3)).
			Return(nil, errors.New("Find() failure"))

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := httpDelete(s.URL + "/queue/queue1/job/3")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("DELETE /queue/$name/job/$id should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		job := &jobqueue.InspectedJob{
			ID:       3,
			Category: "test_job",
			URL:      "http://example.com/",
		}

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			Find(uint64(3)).
			Return(job, nil)
		mockInspector.EXPECT().
			Delete(uint64(3)).
			Return(errors.New("Delete() failure"))

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := httpDelete(s.URL + "/queue/queue1/job/3")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("DELETE /queue/$name/job/$id should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		job := &jobqueue.InspectedJob{
			ID:       3,
			Category: "test_job",
			URL:      "http://example.com/",
		}

		mockInspector := NewMockInspector(ctrl)
		mockInspector.EXPECT().
			Find(uint64(3)).
			Return(job, nil)
		mockInspector.EXPECT().
			Delete(uint64(3)).
			Return(nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			Inspector().
			Return(mockInspector, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := httpDelete(s.URL + "/queue/queue1/job/3")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("DELETE /queue/$name/job/$id should succeed")
		}

		var result jobqueue.InspectedJob
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &result); err != nil {
			t.Error(err)
		}
		if result.ID != job.ID {
			t.Errorf("DELETE /queue/$name/job/$id should return grabbed jobs: %v", result)
		}
	}()
}

func TestGetQueueFailed(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(nil, false)

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("GET /queue/$name/failed should return 404 for an undefinde queue")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(nil, false)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotImplemented {
			t.Error("GET /queue/$name/failed should return 501 if there is no failure log interface")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockFailureLog := NewMockFailureLog(ctrl)
		mockFailureLog.EXPECT().
			FindAllRecentFailures(gomock.Any(), "").
			Return(nil, errors.New("FindAllRecentFailures() failure"))

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(mockFailureLog, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("GET /queue/$name/failed should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		failedJobs := &jobqueue.FailedJobs{
			FailedJobs: []jobqueue.FailedJob{
				{
					ID:       3,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       2,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       1,
					Category: "test_job",
					URL:      "http://example.com/",
				},
			},
			NextCursor: "",
		}

		mockFailureLog := NewMockFailureLog(ctrl)
		mockFailureLog.EXPECT().
			FindAllRecentFailures(gomock.Any(), "").
			Return(failedJobs, nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(mockFailureLog, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name/failed should succeed")
		}

		var failures jobqueue.FailedJobs
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &failures); err != nil {
			t.Error(err)
		}
		if len(failures.FailedJobs) != len(failedJobs.FailedJobs) {
			t.Errorf("GET /queue/$name/failed should return failed jobs: %v", failures)
		}
		for i, f := range failures.FailedJobs {
			if f.ID != failedJobs.FailedJobs[i].ID {
				t.Errorf("GET /queue/$name/failed should return failed jobs: %v", f)
			}
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		failedJobs := &jobqueue.FailedJobs{
			FailedJobs: []jobqueue.FailedJob{
				{
					ID:       3,
					Category: "test_job",
					URL:      "http://example.com/",
					Payload:  json.RawMessage(`{"foo":1}`),
				},
				{
					ID:       2,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       1,
					Category: "test_job",
					URL:      "http://example.com/",
				},
			},
			NextCursor: "",
		}

		limit := uint(123)
		cursor := "foo"

		mockFailureLog := NewMockFailureLog(ctrl)
		mockFailureLog.EXPECT().
			FindAllRecentFailures(limit, cursor).
			Return(failedJobs, nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(mockFailureLog, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed?cursor=" + cursor + "&limit=" + strconv.Itoa(int(limit)))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name/failed should succeed")
		}

		var failures jobqueue.FailedJobs
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &failures); err != nil {
			t.Error(err)
		}
		if len(failures.FailedJobs) != len(failedJobs.FailedJobs) {
			t.Errorf("GET /queue/$name/failed should return failed jobs: %v", failures)
		}
		for i, f := range failures.FailedJobs {
			if f.ID != failedJobs.FailedJobs[i].ID {
				t.Errorf("GET /queue/$name/failed should return failed jobs: %v", f)
			}
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockFailureLog := NewMockFailureLog(ctrl)
		mockFailureLog.EXPECT().
			FindAll(gomock.Any(), "").
			Return(nil, errors.New("FindAll() failure"))

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(mockFailureLog, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed?order=created")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("GET /queue/$name/failed should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		failedJobs := &jobqueue.FailedJobs{
			FailedJobs: []jobqueue.FailedJob{
				{
					ID:       3,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       2,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       1,
					Category: "test_job",
					URL:      "http://example.com/",
				},
			},
			NextCursor: "",
		}

		mockFailureLog := NewMockFailureLog(ctrl)
		mockFailureLog.EXPECT().
			FindAll(gomock.Any(), "").
			Return(failedJobs, nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(mockFailureLog, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed?order=created")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name/failed should succeed")
		}

		var failures jobqueue.FailedJobs
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &failures); err != nil {
			t.Error(err)
		}
		if len(failures.FailedJobs) != len(failedJobs.FailedJobs) {
			t.Errorf("GET /queue/$name/failed should return failed jobs: %v", failures)
		}
		for i, f := range failures.FailedJobs {
			if f.ID != failedJobs.FailedJobs[i].ID {
				t.Errorf("GET /queue/$name/failed should return failed jobs: %v", f)
			}
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		failedJobs := &jobqueue.FailedJobs{
			FailedJobs: []jobqueue.FailedJob{
				{
					ID:       3,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       2,
					Category: "test_job",
					URL:      "http://example.com/",
				},
				{
					ID:       1,
					Category: "test_job",
					URL:      "http://example.com/",
				},
			},
			NextCursor: "",
		}

		limit := uint(123)
		cursor := "foo"

		mockFailureLog := NewMockFailureLog(ctrl)
		mockFailureLog.EXPECT().
			FindAll(limit, cursor).
			Return(failedJobs, nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(mockFailureLog, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed?order=created&cursor=" + cursor + "&limit=" + strconv.Itoa(int(limit)))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name/failed should succeed")
		}

		var failures jobqueue.FailedJobs
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &failures); err != nil {
			t.Error(err)
		}
		if len(failures.FailedJobs) != len(failedJobs.FailedJobs) {
			t.Errorf("GET /queue/$name/failed should return failed jobs: %v", failures)
		}
		for i, f := range failures.FailedJobs {
			if f.ID != failedJobs.FailedJobs[i].ID {
				t.Errorf("GET /queue/$name/failed should return failed jobs: %v", f)
			}
		}
	}()
}

func TestGetQueueFailedJob(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, _ := newMockServer(ctrl)
		defer s.Close()

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed/a")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Error("GET /queue/$name/failed/$id should reject invalid ID")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(nil, false)

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed/5")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("GET /queue/$name/failed/$id should return 404 for an undefinde queue")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(nil, false)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed/5")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotImplemented {
			t.Error("GET /queue/$name/failed/$id should return 501 if there is no failure log interface")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockFailureLog := NewMockFailureLog(ctrl)
		mockFailureLog.EXPECT().
			Find(uint64(5)).
			Return(nil, sql.ErrNoRows)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(mockFailureLog, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed/5")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("GET /queue/$name/failed/$id should return 404 for an unknown job")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockFailureLog := NewMockFailureLog(ctrl)
		mockFailureLog.EXPECT().
			Find(uint64(5)).
			Return(nil, errors.New("Find() failure"))

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(mockFailureLog, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed/5")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("GET /queue/$name/failed/$id should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		failedJob := &jobqueue.FailedJob{
			ID:       5,
			JobID:    3,
			Category: "test_job",
			URL:      "http://example.com/",
		}

		mockFailureLog := NewMockFailureLog(ctrl)
		mockFailureLog.EXPECT().
			Find(uint64(5)).
			Return(failedJob, nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(mockFailureLog, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := http.Get(s.URL + "/queue/failed_queue/failed/5")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /queue/$name/failed/$id should succeed")
		}

		var failure jobqueue.FailedJob
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &failure); err != nil {
			t.Error(err)
		}
		if failure.ID != failedJob.ID {
			t.Errorf("GET /queue/$name/failed/$id should return failed jobs: %v", failure)
		}
	}()
}

func TestDeleteQueueFailedJob(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, _ := newMockServer(ctrl)
		defer s.Close()

		resp, err := httpDelete(s.URL + "/queue/failed_queue/failed/a")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Error("DELETE /queue/$name/failed/$id should reject invalid ID")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(nil, false)

		resp, err := httpDelete(s.URL + "/queue/failed_queue/failed/5")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("DELETE /queue/$name/failed/$id should return 404 for an undefinde queue")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(nil, false)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := httpDelete(s.URL + "/queue/failed_queue/failed/5")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotImplemented {
			t.Error("DELETE /queue/$name/failed/$id should return 501 if there is no failure log interface")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockFailureLog := NewMockFailureLog(ctrl)
		mockFailureLog.EXPECT().
			Find(uint64(5)).
			Return(nil, sql.ErrNoRows)
		mockFailureLog.EXPECT().
			Delete(uint64(5)).
			Return(nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(mockFailureLog, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := httpDelete(s.URL + "/queue/failed_queue/failed/5")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("DELETE /queue/$name/failed/$id should return 404 for an unknown job")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockFailureLog := NewMockFailureLog(ctrl)
		mockFailureLog.EXPECT().
			Find(uint64(5)).
			Return(nil, errors.New("Find() failure"))

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(mockFailureLog, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := httpDelete(s.URL + "/queue/failed_queue/failed/5")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("DELETE /queue/$name/failed/$id should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		failedJob := &jobqueue.FailedJob{
			ID:       5,
			JobID:    3,
			Category: "test_job",
			URL:      "http://example.com/",
		}

		mockFailureLog := NewMockFailureLog(ctrl)
		mockFailureLog.EXPECT().
			Find(uint64(5)).
			Return(failedJob, nil)
		mockFailureLog.EXPECT().
			Delete(uint64(5)).
			Return(errors.New("Delete() failure"))

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(mockFailureLog, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := httpDelete(s.URL + "/queue/failed_queue/failed/5")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("DELETE /queue/$name/failed/$id should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		failedJob := &jobqueue.FailedJob{
			ID:       5,
			JobID:    3,
			Category: "test_job",
			URL:      "http://example.com/",
		}

		mockFailureLog := NewMockFailureLog(ctrl)
		mockFailureLog.EXPECT().
			Find(uint64(5)).
			Return(failedJob, nil)
		mockFailureLog.EXPECT().
			Delete(uint64(5)).
			Return(nil)

		mockJobQueue := NewMockJobQueue(ctrl)
		mockJobQueue.EXPECT().
			FailureLog().
			Return(mockFailureLog, true)

		mockApp.Service.EXPECT().
			GetJobQueue(gomock.Any()).
			Return(newMockRunningQueue(mockJobQueue, nil), true)

		resp, err := httpDelete(s.URL + "/queue/failed_queue/failed/5")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("DELETE /queue/$name/failed/$id should succeed")
		}

		var failure jobqueue.FailedJob
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &failure); err != nil {
			t.Error(err)
		}
		if failure.ID != failedJob.ID {
			t.Errorf("DELETE /queue/$name/failed/$id should return failed jobs: %v", failure)
		}
	}()
}

type jobQueue = jobqueue.JobQueue

type mockRunningQueue struct {
	jobQueue
	stats *dispatcher.Stats
}

func newMockRunningQueue(jq jobQueue, stats *dispatcher.Stats) *mockRunningQueue {
	return &mockRunningQueue{jq, stats}
}

func (q *mockRunningQueue) PollingInterval() uint {
	return 0
}

func (q *mockRunningQueue) MaxWorkers() uint {
	return 0
}

func (q *mockRunningQueue) WorkerStats() *dispatcher.Stats {
	return q.stats
}
