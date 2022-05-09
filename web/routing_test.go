package web

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/fireworq/fireworq/model"
	"github.com/fireworq/fireworq/repository"

	"github.com/golang/mock/gomock"
)

func TestGetRoutingList(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.RoutingRepository.EXPECT().
			FindAll().
			Return([]model.Routing{}, errors.New("FindAll() failure"))

		resp, err := http.Get(s.URL + "/routings")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("GET /routings should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.RoutingRepository.EXPECT().
			FindAll().
			Return([]model.Routing{}, nil)

		resp, err := http.Get(s.URL + "/routings")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /routings should succeed")
		}

		var rs []model.Routing
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &rs); err != nil {
			t.Error("GET /routings should return routing definitions")
		}
		if len(rs) != 0 {
			t.Error("GET /routings should return empty list if nothing defined")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		routings := []model.Routing{
			{
				QueueName:   "queue1",
				JobCategory: "job1",
			},
			{
				QueueName:   "queue2",
				JobCategory: "job2",
			},
			{
				QueueName:   "queue3",
				JobCategory: "job3",
			},
		}
		mockApp.RoutingRepository.EXPECT().
			FindAll().
			Return(routings, nil)

		resp, err := http.Get(s.URL + "/routings")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /routings should succeed")
		}

		var rs []model.Routing
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &rs); err != nil {
			t.Error(err)
		}
		if len(rs) != len(routings) {
			t.Error("GET /routings should return defined routings")
		}
		for i, r := range rs {
			if r.QueueName != routings[i].QueueName || r.JobCategory != routings[i].JobCategory {
				t.Error("GET /routings should return defined routings")
			}
		}
	}()
}

func TestGetRouting(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.RoutingRepository.EXPECT().
			FindQueueNameByJobCategory("job1").
			Return("")

		resp, err := http.Get(s.URL + "/routing/job1")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("GET /routing/$category should return 404")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.RoutingRepository.EXPECT().
			FindQueueNameByJobCategory("job2").
			Return("queue2")

		resp, err := http.Get(s.URL + "/routing/job2")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("GET /routing/$category should succeed")
		}

		var r model.Routing
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &r); err != nil {
			t.Error(err)
		}
		if r.JobCategory != "job2" || r.QueueName != "queue2" {
			t.Error("GET /routing/$name should return a defined routing")
		}
	}()
}

func TestPutRouting(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, _ := newMockServer(ctrl)
		defer s.Close()

		resp, err := putJSON(s.URL+"/routing/job3", "foo")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Error("PUT /routing/$name should reject invalid input")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.RoutingRepository.EXPECT().
			Add("job3", "queue3").
			Return(false, &repository.QueueNotFoundError{QueueName: "queue3"})

		def := &model.Routing{QueueName: "queue3"}

		resp, err := putJSON(s.URL+"/routing/job3", def)
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("PUT /routing/$category should not accept undefined queue")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.RoutingRepository.EXPECT().
			Add("job3", "queue3").
			Return(false, errors.New("Add() failure"))

		def := &model.Routing{QueueName: "queue3"}

		resp, err := putJSON(s.URL+"/routing/job3", def)
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("PUT /routing/$category should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		def := &model.Routing{QueueName: "queue4", JobCategory: "job4"}

		mockApp.RoutingRepository.EXPECT().
			Add(def.JobCategory, def.QueueName).
			Return(true, nil)

		resp, err := putJSON(s.URL+"/routing/job4", def)
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("PUT /routing/$category should succeed")
		}

		var r model.Routing
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &r); err != nil {
			t.Error(err)
		}
		if r.JobCategory != def.JobCategory || r.QueueName != def.QueueName {
			t.Error("PUT /routing/$name should return a defined routing")
		}
	}()
}

func TestDeleteRouting(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.RoutingRepository.EXPECT().
			FindQueueNameByJobCategory("job1").
			Return("")

		resp, err := httpDelete(s.URL + "/routing/job1")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("DELETE /routing/$category should return 404")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.RoutingRepository.EXPECT().
			FindQueueNameByJobCategory("job2").
			Return("queue2")
		mockApp.RoutingRepository.EXPECT().
			DeleteByJobCategory("job2").
			Return(errors.New("DeleteByJobCategory() failure"))

		resp, err := httpDelete(s.URL + "/routing/job2")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("DELETE /routing/$category should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.RoutingRepository.EXPECT().
			FindQueueNameByJobCategory("job2").
			Return("queue2")
		mockApp.RoutingRepository.EXPECT().
			DeleteByJobCategory("job2").
			Return(nil)

		resp, err := httpDelete(s.URL + "/routing/job2")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("DELETE /routing/$category should succeed")
		}

		var r model.Routing
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &r); err != nil {
			t.Error(err)
		}
		if r.JobCategory != "job2" || r.QueueName != "queue2" {
			t.Error("DELETE /routing/$name should return a deleted routing")
		}
	}()
}
