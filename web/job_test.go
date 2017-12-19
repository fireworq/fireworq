package web

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/fireworq/fireworq/service"

	"github.com/golang/mock/gomock"
)

func TestPostJob(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, _ := newMockServer(ctrl)
		defer s.Close()

		func() {
			resp, err := http.Get(s.URL + "/job/test_job0")
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("/job/$category should only accept POST method")
			}
		}()

		func() {
			resp, err := putJSON(s.URL+"/job/test_job0", emptyObject{})
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Error("/job/$category should only accept POST method")
			}
		}()
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, _ := newMockServer(ctrl)
		defer s.Close()

		func() {
			resp, err := http.Post(s.URL+"/job/test_job0", "application/json", strings.NewReader(`{"url":"http://example.com/","payload":{},"run_after":"foo"}`))
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Error("POST /job/$category should reject invalid input")
			}
		}()

		func() {
			resp, err := http.Post(s.URL+"/job/test_job0", "application/json", strings.NewReader(`{"url":"http://example.com/","payload":aaa}`))
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Error("POST /job/$category should reject invalid input")
			}
		}()

		func() {
			resp, err := http.Post(s.URL+"/job/test_job0", "application/json", strings.NewReader(`{"url":"http://example.com/","payload":"\u"}`))
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Error("POST /job/$category should reject invalid input")
			}
		}()

		func() {
			resp, err := http.Post(s.URL+"/job/test_job0", "application/json", strings.NewReader(`{"payload":{}}`))
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Error("POST /job/$category should reject invalid input")
			}
		}()
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		mockApp.Service.EXPECT().
			Push(gomock.Any()).
			Return(nil, errors.New("Push() failure"))

		resp, err := http.Post(s.URL+"/job/test_job1", "application/json", strings.NewReader(`{"url":"http://example.com/","payload":{}}`))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Error("POST /job/$category should fail")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		job := &IncomingJob{
			CategoryField: "test_job1",
			URLField:      "http://example.com/",
			PayloadField:  json.RawMessage("{}"),
		}

		result := &service.PushResult{
			ID:        123,
			QueueName: "default",
		}

		mockApp.Service.EXPECT().
			Push(gomock.Any()).
			Return(result, nil)

		resp, err := http.Post(s.URL+"/job/test_job1", "application/json", strings.NewReader(`{"url":"http://example.com/","payload":{}}`))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("POST /job/$category should accept a job")
		}

		var j PushResult
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &j); err != nil {
			t.Error("POST /job/$category should return a pushed job")
		}
		if j.ID != 123 || j.QueueName != "default" || j.CategoryField != job.CategoryField || j.URLField != job.URLField || j.Payload() != job.Payload() {
			t.Error("POST /job/$category should return a pushed job")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		job := &IncomingJob{
			CategoryField: "test_job2",
			URLField:      "http://example.com/",
			PayloadField:  json.RawMessage(`"foo bar"`),
		}

		result := &service.PushResult{
			ID:        124,
			QueueName: "default",
		}

		mockApp.Service.EXPECT().
			Push(gomock.Any()).
			Return(result, nil)

		resp, err := http.Post(s.URL+"/job/test_job2", "application/json", strings.NewReader(`{"url":"http://example.com/","payload":"foo bar"}`))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("POST /job/$category should accept a job")
		}

		var j PushResult
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &j); err != nil {
			t.Error("POST /job/$category should return a pushed job")
		}
		if j.Payload() != "foo bar" {
			t.Error("POST /job/$category should accept an arbitrary payload")
		}
		if j.ID != 124 || j.QueueName != "default" || j.CategoryField != job.CategoryField || j.URLField != job.URLField || j.Payload() != job.Payload() {
			t.Error("POST /job/$category should return a pushed job")
		}
	}()

	func() {
		ctrl := gomock.NewController(t)
		s, mockApp := newMockServer(ctrl)
		defer s.Close()

		job := &IncomingJob{
			CategoryField: "test_job3",
			URLField:      "http://example.com/",
			PayloadField:  json.RawMessage(""),
		}

		result := &service.PushResult{
			ID:        1,
			QueueName: "queue1",
		}

		mockApp.Service.EXPECT().
			Push(gomock.Any()).
			Return(result, nil)

		resp, err := http.Post(s.URL+"/job/test_job3", "application/json", strings.NewReader(`{"url":"http://example.com/"}`))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Error("POST /job/$category should accept a job")
		}

		var j PushResult
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &j); err != nil {
			t.Error("POST /job/$category should return a pushed job")
		}
		if j.Payload() != "" {
			t.Error("POST /job/$category should accept an empty payload")
		}
		if j.ID != 1 || j.QueueName != "queue1" || j.CategoryField != job.CategoryField || j.URLField != job.URLField || j.Payload() != job.Payload() {
			t.Error("POST /job/$category should return a pushed job")
		}
	}()
}

func TestIncomingJob(t *testing.T) {
	{
		j := &IncomingJob{
			CategoryField: "test_job",
			URLField:      "http://example.com/",
			PayloadField:  json.RawMessage(`"foo"`),
		}

		if j.Category() != "test_job" {
			t.Error("Wrong job property")
		}
		if j.URL() != "http://example.com/" {
			t.Error("Wrong job property")
		}
		if j.Payload() != "foo" {
			t.Error("Wrong job property")
		}
		if j.NextDelay() != 0 {
			t.Error("Wrong job property")
		}
		if j.RetryCount() != 0 {
			t.Error("Wrong job property")
		}
		if j.RetryDelay() != 0 {
			t.Error("Wrong job property")
		}
		if j.Timeout() != 0 {
			t.Error("Wrong job property")
		}
	}

	{
		j := &IncomingJob{
			CategoryField: "test_job",
			URLField:      "http://example.com/",
			PayloadField:  json.RawMessage(`"\a"`),
		}

		if j.Category() != "test_job" {
			t.Error("Wrong job property")
		}
		if j.URL() != "http://example.com/" {
			t.Error("Wrong job property")
		}
		if j.Payload() != "" {
			t.Error("Wrong job property")
		}
		if j.NextDelay() != 0 {
			t.Error("Wrong job property")
		}
		if j.RetryCount() != 0 {
			t.Error("Wrong job property")
		}
		if j.RetryDelay() != 0 {
			t.Error("Wrong job property")
		}
		if j.Timeout() != 0 {
			t.Error("Wrong job property")
		}
	}

	{
		j := &IncomingJob{
			CategoryField:   "test_job",
			URLField:        "http://example.com/",
			PayloadField:    json.RawMessage("{}"),
			RunAfterField:   60,
			MaxRetriesField: 5,
			RetryDelayField: 120,
			TimeoutField:    10,
		}

		if j.Category() != "test_job" {
			t.Error("Wrong job property")
		}
		if j.URL() != "http://example.com/" {
			t.Error("Wrong job property")
		}
		if j.Payload() != "{}" {
			t.Error("Wrong job property")
		}
		if j.NextDelay() != uint64(60000) {
			t.Error("Wrong job property")
		}
		if j.RetryCount() != uint(5) {
			t.Error("Wrong job property")
		}
		if j.RetryDelay() != uint(120) {
			t.Error("Wrong job property")
		}
		if j.Timeout() != uint(10) {
			t.Error("Wrong job property")
		}
	}
}
