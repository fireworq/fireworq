package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/fireworq/fireworq/config"

	"github.com/golang/mock/gomock"
)

func TestGetVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	s, _ := newMockServer(ctrl)
	defer s.Close()

	resp, err := http.Get(s.URL + "/version")
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if string(body) != "Fireworq 0.1.0-TEST\n" {
		t.Error("GET /version should return a version string")
	}
}

func TestGetSettings(t *testing.T) {
	func() {
		ctrl := gomock.NewController(t)
		s, _ := newMockServer(ctrl)
		defer s.Close()

		resp, err := http.Get(s.URL + "/settings")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()

		var settings map[string]string
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &settings); err != nil {
			t.Error("GET /settings should return settings")
		}

		if len(settings) != len(config.Keys()) {
			t.Error("GET /settings should return all the settings")
		}
	}()

	config.Locally("driver", "broken", func() {
		ctrl := gomock.NewController(t)
		s, _ := newMockServer(ctrl)
		defer s.Close()

		resp, err := http.Get(s.URL + "/settings")
		if err != nil {
			t.Error(err)
		}
		defer resp.Body.Close()

		var settings map[string]string
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(buf, &settings); err != nil {
			t.Error("GET /settings should return settings")
		}

		if len(settings) != len(config.Keys()) {
			t.Error("GET /settings should return all the settings")
		}

		if settings["driver"] != "broken" {
			t.Error("GET /settings should return a modified setting")
		}
	})
}
