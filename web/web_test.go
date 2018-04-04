//go:generate mockgen -source application.go -destination mock_web_test.go -package web
//go:generate mockgen -source ../repository/repository.go -destination mock_web_repository_test.go -package web
//go:generate mockgen -source ../jobqueue/jobqueue.go -destination mock_jobqueue_test.go -package web -self_package github.com/fireworq/fireworq/web
//go:generate mockgen -source ../jobqueue/inspector.go -destination mock_inspector_test.go -package web -self_package github.com/fireworq/fireworq/web

package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestMain(m *testing.M) {
	Init()
	transport := http.DefaultTransport.(*http.Transport)
	transport.DisableKeepAlives = disableKeepAlives

	os.Exit(m.Run())
}

func postJSON(url string, value interface{}) (*http.Response, error) {
	j, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return http.Post(url, "application/json", bytes.NewBuffer(j))
}

func putJSON(url string, value interface{}) (*http.Response, error) {
	j, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	req.ContentLength = int64(len(j))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	return client.Do(req)
}

func httpDelete(url string) (*http.Response, error) {
	buf := make([]byte, 0)
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	return client.Do(req)
}

func NewMockApplication(ctrl *gomock.Controller) *Application {
	mockQueueRepo := NewMockQueueRepository(ctrl)
	mockRoutingRepo := NewMockRoutingRepository(ctrl)
	mockService := NewMockService(ctrl)
	return &Application{
		Service:           mockService,
		QueueRepository:   mockQueueRepo,
		RoutingRepository: mockRoutingRepo,
		Version:           "Fireworq 0.1.0-TEST",
	}
}

func newMockServer(ctrl *gomock.Controller) (*httptest.Server, *mockApplication) {
	app := NewMockApplication(ctrl)
	return httptest.NewServer(app.newServer().mux), &mockApplication{
		Service:           app.Service.(*MockService),
		QueueRepository:   app.QueueRepository.(*MockQueueRepository),
		RoutingRepository: app.RoutingRepository.(*MockRoutingRepository),
	}
}

type mockApplication struct {
	Service           *MockService
	QueueRepository   *MockQueueRepository
	RoutingRepository *MockRoutingRepository
}

type emptyObject struct{}
