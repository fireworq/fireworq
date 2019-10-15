package worker

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/fireworq/fireworq/config"
	"github.com/fireworq/fireworq/jobqueue"
	"github.com/fireworq/fireworq/jobqueue/logger"
)

func TestMain(m *testing.M) {
	HTTPInit()
	os.Exit(m.Run())
}

func TestHTTPInit(t *testing.T) {
	config.Locally("dispatch_max_conns_per_host", "foo", func() {
		HTTPInit()

		transport := http.DefaultTransport.(*http.Transport)
		if transport.MaxIdleConnsPerHost != 10 {
			t.Error("Not set a default value to MaxIdleConnsPerHost")
		}
	})

	config.Locally("dispatch_max_conns_per_host", "1000", func() {
		HTTPInit()

		transport := http.DefaultTransport.(*http.Transport)
		if transport.MaxIdleConnsPerHost != 1000 {
			t.Error("Not set to MaxIdleConnsPerHost")
		}
	})

	config.Locally("dispatch_idle_conn_timeout", "foo", func() {
		HTTPInit()

		transport := http.DefaultTransport.(*http.Transport)
		if transport.IdleConnTimeout != 0 {
			t.Error("Not set a default value to IdleConnTimeout")
		}
	})

	config.Locally("dispatch_idle_conn_timeout", "10", func() {
		HTTPInit()

		transport := http.DefaultTransport.(*http.Transport)
		if transport.IdleConnTimeout != 10*time.Second {
			t.Error("Not set to IdleConnTimeout")
		}
	})
}

func TestNewWorker(t *testing.T) {
	func() {
		w0 := &HTTPWorker{}
		w1, ok := w0.NewWorker().(*HTTPWorker)
		if !ok {
			t.Error("NewWorker should create a new instance")
		}
		if w1.Logger == nil {
			t.Error("A new worker should have a logger")
		}
	}()

	func() {
		w0 := &HTTPWorker{
			UserAgent: "foo",
		}

		w1, ok := w0.NewWorker().(*HTTPWorker)
		if !ok {
			t.Error("NewWorker should create a new instance")
		}
		if w1.UserAgent != w0.UserAgent {
			t.Error("A new worker should inherit user agent")
		}

		w1.UserAgent = "bar"
		if w1.UserAgent == w0.UserAgent {
			t.Error("Modifying user agent should not affect the ancestor")
		}
	}()
}

func TestWork(t *testing.T) {
	server := newTestWorker(t)
	defer server.close()

	ua := "TestUserAgent/1.0"
	wc := &HTTPWorker{
		UserAgent: ua,
	}
	w := wc.NewWorker()

	func() {
		payload := `{"status":"success"}`
		rslt := w.Work(&job{
			url:     "",
			payload: payload,
		})
		if rslt.Status != jobqueue.ResultStatusInternalFailure {
			t.Errorf("Worker request should fail internally")
		}
	}()

	func() {
		payload := `{"status":"success"}`
		rslt := w.Work(&job{
			url:     ":",
			payload: payload,
		})
		if rslt.Status != jobqueue.ResultStatusInternalFailure {
			t.Errorf("Worker request should fail internally")
		}
	}()

	func() {
		payload := `{"status":"success"}`
		rslt := w.Work(&job{
			url:     server.url(),
			payload: payload,
		})
		if rslt.IsFailure() {
			t.Errorf("Worker request should succeed")
		}

		server.wait(1 * time.Second)
		if server.payload() != payload {
			t.Errorf("Wrong payload '%s' sent to the server", server.payload())
		}
		if server.ua() != ua {
			t.Errorf("Wrong UA: %s", server.ua())
		}
	}()

	func() {
		payload := `{"status":"success","message":"foo bar"}`
		rslt := w.Work(&job{
			url:     server.url(),
			payload: payload,
		})
		if rslt.IsFailure() {
			t.Errorf("A message should be accepted")
		}

		server.wait(1 * time.Second)
		if server.payload() != payload {
			t.Errorf("Wrong payload '%s' sent to the server", server.payload())
		}
		if server.ua() != ua {
			t.Errorf("Wrong UA: %s", server.ua())
		}
	}()

	func() {
		payload := `{"status":"failure"}`
		rslt := w.Work(&job{
			url:     server.url(),
			payload: payload,
		})
		if rslt.Status != jobqueue.ResultStatusFailure {
			t.Errorf("Worker request should fail")
		}

		server.wait(1 * time.Second)
		if server.payload() != payload {
			t.Errorf("Wrong payload '%s' sent to the server", server.payload())
		}
		if server.ua() != ua {
			t.Errorf("Wrong UA: %s", server.ua())
		}
	}()

	func() {
		payload := `{"status":"permanent-failure"}`
		rslt := w.Work(&job{
			url:     server.url(),
			payload: payload,
		})
		if rslt.Status != jobqueue.ResultStatusPermanentFailure {
			t.Errorf("Worker request should permanently fail")
		}

		server.wait(1 * time.Second)
		if server.payload() != payload {
			t.Errorf("Wrong payload '%s' sent to the server", server.payload())
		}
		if server.ua() != ua {
			t.Errorf("Wrong UA: %s", server.ua())
		}
	}()

	func() {
		payload := `"foo bar"`
		rslt := w.Work(&job{
			url:     server.url(),
			payload: payload,
		})
		if rslt.Status != jobqueue.ResultStatusFailure {
			t.Errorf("Worker request should fail for illegal response")
		}

		server.wait(1 * time.Second)
		if server.payload() != payload {
			t.Errorf("Wrong payload '%s' sent to the server", server.payload())
		}
		if server.ua() != ua {
			t.Errorf("Wrong UA: %s", server.ua())
		}
	}()

	func() {
		payload := `{"status": "ok"}`
		rslt := w.Work(&job{
			url:     server.url(),
			payload: payload,
		})
		if rslt.Status != jobqueue.ResultStatusFailure {
			t.Errorf("Worker request should fail for illegal response")
		}

		server.wait(1 * time.Second)
		if server.payload() != payload {
			t.Errorf("Wrong payload '%s' sent to the server", server.payload())
		}
		if server.ua() != ua {
			t.Errorf("Wrong UA: %s", server.ua())
		}
	}()

	func() {
		payload := `{"status":`
		rslt := w.Work(&job{
			url:     server.url(),
			payload: payload,
		})
		if rslt.Status != jobqueue.ResultStatusFailure {
			t.Errorf("Worker request should fail for broken response")
		}

		server.wait(1 * time.Second)
		if server.payload() != payload {
			t.Errorf("Wrong payload '%s' sent to the server", server.payload())
		}
		if server.ua() != ua {
			t.Errorf("Wrong UA: %s", server.ua())
		}
	}()

	func() {
		payload := `{"status":"success"}`
		wc := &HTTPWorker{}
		w := wc.NewWorker()
		rslt := w.Work(&job{
			url:     server.url(),
			payload: payload,
		})
		if rslt.Status != jobqueue.ResultStatusSuccess {
			t.Errorf("Worker request should succeed")
		}

		server.wait(1 * time.Second)
		if server.payload() != payload {
			t.Errorf("Wrong payload '%s' sent to the server", server.payload())
		}
		if server.ua() != config.Get("dispatch_user_agent") {
			t.Errorf("Wrong UA: %s", server.ua())
		}
	}()
}

type testServer struct {
	worker *testWorker
	server *httptest.Server
	t      *testing.T
}

func newTestWorker(t *testing.T) *testServer {
	w := &testWorker{request: make(chan struct{}, 1)}
	return &testServer{
		worker: w,
		server: httptest.NewServer(w),
		t:      t,
	}
}

func (s *testServer) url() string {
	return s.server.URL
}

func (s *testServer) payload() string {
	return s.worker.payload
}

func (s *testServer) ua() string {
	return s.worker.ua
}

func (s *testServer) wait(dur time.Duration) {
	select {
	case <-s.worker.request:
	case <-time.After(dur):
		s.t.Error("Timeout")
	}
}

func (s *testServer) close() {
	s.server.Close()
}

type testWorker struct {
	payload string
	ua      string
	request chan struct{}
}

func (worker *testWorker) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	worker.ua = req.Header.Get("User-Agent")

	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(500)
		worker.payload = err.Error()
	} else {
		w.WriteHeader(200)
		worker.payload = string(buf)
	}

	w.Write(buf)
	worker.request <- struct{}{}
}

type job struct {
	url     string
	payload string
}

func (j *job) URL() string                    { return j.url }
func (j *job) Payload() string                { return j.payload }
func (j *job) RetryCount() uint               { return 0 }
func (j *job) RetryDelay() uint               { return 0 }
func (j *job) FailCount() uint                { return 0 }
func (j *job) Timeout() uint                  { return 0 }
func (j *job) ToLoggable() logger.LoggableJob { return nil }
