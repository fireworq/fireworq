package web

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/fireworq/fireworq/config"

	serverstarter "github.com/lestrrat/go-server-starter/listener"
)

func testServer(t *testing.T, addr string) {
	s := newServer(os.Stdout)
	s.handle("/test", serveTest)

	server, err := s.start()
	if err != nil {
		t.Error(err)
	}
	defer server.Close()

	if len(s.addrs) != 1 || s.addrs[0].Network() != "tcp" {
		t.Error("Must listen on one TCP port")
	}

	if addr != "" && addr != s.addrs[0].String() {
		t.Errorf("Wrong address: %s != %s", s.addrs[0].String(), addr)
	}

	time.Sleep(1 * time.Second)

	func() {
		resp, err := http.Get(fmt.Sprintf("http://%s/test", s.addrs[0].String()))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}

		if string(body) != testContent {
			t.Errorf("Wrong content: %s", string(body))
		}
	}()

	func() {
		resp, err := http.Get(fmt.Sprintf("http://%s/notfound", s.addrs[0].String()))
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Error("Wrong status")
		}
	}()
}

func TestServerStart(t *testing.T) {
	config.Locally("bind", "127.0.0.1:0", func() {
		testServer(t, "")
	})

	config.Locally("bind", "127.0.0.1:1234", func() {
		s1 := newServer(os.Stdout)
		s1.handle("/test", serveTest)

		server1, _ := s1.start()
		defer func() {
			if server1 != nil {
				server1.Close()
			}
		}()

		time.Sleep(1 * time.Second) // wait for up

		s2 := newServer(os.Stdout)
		s2.handle("/test", serveTest)

		if server2, err := s2.start(); err == nil {
			server2.Close()
			t.Error("It should fail starting on the same port")
		}
	})
}

func TestServerStarter(t *testing.T) {
	func() {
		original := os.Getenv(serverstarter.ServerStarterEnvVarName)
		defer func() {
			os.Setenv(serverstarter.ServerStarterEnvVarName, original)
		}()

		os.Setenv(serverstarter.ServerStarterEnvVarName, "=x")

		s := newServer(os.Stdout)
		s.handle("/test", serveTest)

		if server, err := s.start(); err == nil {
			server.Close()
			t.Error("It should fail starting on an invalid SERVER_STARTER_PORT")
		}
	}()
}

func TestGraceful(t *testing.T) {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Error(err)
	}

	// Shutdown during server idle
	config.Locally("bind", "127.0.0.1:0", func() {
		s := newServer(os.Stdout)
		s.handle("/test", serveTest)

		server, err := s.start()
		if err != nil {
			t.Error(err)
		}
		if len(s.addrs) != 1 || s.addrs[0].Network() != "tcp" {
			t.Error("Must listen on one TCP port")
		}

		stopped := make(chan struct{})
		go func() {
			left, err := graceful(server, 1*time.Second)
			if err != nil {
				t.Error(err)
			}
			if left <= 0 {
				t.Error("There should be some time left")
			}
			stopped <- struct{}{}
		}()

		time.Sleep(1 * time.Second) // wait for up

		if err := p.Signal(syscall.SIGTERM); err != nil {
			t.Error(err)
		}

		select {
		case <-time.After(2 * time.Second):
			t.Error("Timed out")
		case <-stopped:
		}
	})

	// Shutdown during server request handling and there is enough
	// time to respond
	config.Locally("bind", "127.0.0.1:0", func() {
		s := newServer(os.Stdout)
		h := newSleepHandler()
		s.handle("/sleep", h.serveSleep)

		server, err := s.start()
		if err != nil {
			t.Error(err)
		}
		if len(s.addrs) != 1 || s.addrs[0].Network() != "tcp" {
			t.Error("Must listen on one TCP port")
		}

		stopped := make(chan struct{})
		go func() {
			left, err := graceful(server, 10*time.Second)
			if err != nil {
				t.Error(err)
			}
			if left < 6*time.Second {
				t.Error("There should be some time left")
			}
			stopped <- struct{}{}
		}()

		time.Sleep(1 * time.Second) // wait for up

		fetched := make(chan struct{})
		go func() {
			resp, err := http.Get(fmt.Sprintf("http://%s/sleep", s.addrs[0].String()))
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Error(err)
			}
			if string(body) != testContent {
				t.Errorf("Wrong content: %s", string(body))
			}
			fetched <- struct{}{}
		}()
		select {
		case <-time.After(3 * time.Second):
			t.Error("Timed out")
		case <-h.requested:
		}

		if err := p.Signal(syscall.SIGTERM); err != nil {
			t.Error(err)
		}

		select {
		case <-time.After(4 * time.Second):
			t.Error("Timed out")
		case <-stopped:
		}

		select {
		case <-time.After(2 * time.Second):
			t.Error("Request failed")
		case <-fetched:
		}
	})

	// Shutdown during server request handling and there is no enough
	// time to respond
	config.Locally("bind", "127.0.0.1:0", func() {
		s := newServer(os.Stdout)
		h := newSleepHandler()
		s.handle("/sleep", h.serveSleep)

		server, err := s.start()
		if err != nil {
			t.Error(err)
		}
		if len(s.addrs) != 1 || s.addrs[0].Network() != "tcp" {
			t.Error("Must listen on one TCP port")
		}

		stopped := make(chan struct{})
		go func() {
			left, err := graceful(server, 1*time.Second)
			if err != context.DeadlineExceeded {
				t.Errorf("Wrong error: %v", err)
			}
			if left > 0 {
				t.Errorf("There should be no time left: %v", left)
			}
			stopped <- struct{}{}
		}()

		time.Sleep(1 * time.Second) // wait for up

		go func() {
			resp, _ := http.Get(fmt.Sprintf("http://%s/sleep", s.addrs[0].String()))
			if resp != nil {
				resp.Body.Close()
			}
		}()
		select {
		case <-time.After(3 * time.Second):
			t.Error("Timed out")
		case <-h.requested:
		}

		if err := p.Signal(syscall.SIGTERM); err != nil {
			t.Error(err)
		}

		select {
		case <-time.After(2 * time.Second):
			t.Error("Timed out")
		case <-stopped:
		}
	})

	// Shutdown during server request handling and there is enough
	// time to respond but it is forced to shut down
	config.Locally("bind", "127.0.0.1:0", func() {
		s := newServer(os.Stdout)
		h := newSleepHandler()
		s.handle("/sleep", h.serveSleep)

		server, err := s.start()
		if err != nil {
			t.Error(err)
		}
		if len(s.addrs) != 1 || s.addrs[0].Network() != "tcp" {
			t.Error("Must listen on one TCP port")
		}

		stopped := make(chan struct{})
		go func() {
			left, err := graceful(server, 10*time.Second)
			if err != nil {
				t.Error(err)
			}
			if left != 0 {
				t.Error("There should be no time left")
			}
			stopped <- struct{}{}
		}()

		time.Sleep(1 * time.Second) // wait for up

		go func() {
			resp, _ := http.Get(fmt.Sprintf("http://%s/sleep", s.addrs[0].String()))
			if resp != nil {
				resp.Body.Close()
			}
		}()
		select {
		case <-time.After(3 * time.Second):
			t.Error("Timed out")
		case <-h.requested:
		}

		if err := p.Signal(syscall.SIGTERM); err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond)
		if err := p.Signal(syscall.SIGTERM); err != nil {
			t.Error(err)
		}

		select {
		case <-time.After(2 * time.Second):
			t.Error("Timed out")
		case <-stopped:
		}
	})
}

func serveTest(w http.ResponseWriter, req *http.Request) error {
	fmt.Fprint(w, testContent)
	return nil
}

type sleepHandler struct {
	requested chan struct{}
}

func newSleepHandler() *sleepHandler {
	return &sleepHandler{make(chan struct{}, 1)}
}

func (h *sleepHandler) serveSleep(w http.ResponseWriter, req *http.Request) error {
	h.requested <- struct{}{}
	time.Sleep(3 * time.Second)
	fmt.Fprint(w, testContent)
	return nil
}

const testContent = "This is a test\n"
