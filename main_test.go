package main

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/fireworq/fireworq/config"
	"github.com/fireworq/fireworq/test"
)

func TestMain(m *testing.M) {
	initTestMain()

	status, err := test.Run(m)
	if err != nil {
		panic(err)
	}
	os.Exit(status)
}

func TestParseCmdArgs(t *testing.T) {
	{
		_, err := parseCmdArgs([]string{"--hoge"})
		if err == nil {
			t.Error("Should not accept unknown option")
		}
	}
	{
		args, err := parseCmdArgs([]string{"-v"})
		if err != nil {
			t.Error(err)
		}
		if !args.showVersion {
			t.Error("-v must show version")
		}
	}
	{
		args, err := parseCmdArgs([]string{"--version"})
		if err != nil {
			t.Error(err)
		}
		if !args.showVersion {
			t.Error("--version must show version")
		}
	}
	{
		args, err := parseCmdArgs([]string{"--credits"})
		if err != nil {
			t.Error(err)
		}
		if !args.showCredits {
			t.Error("--credits must show credits")
		}
	}
	{
		args, err := parseCmdArgs([]string{"--pid=fireworq.pid", "--bind", "0.0.0.0:80", "--error-log-level", "warn"})
		if err != nil {
			t.Error(err)
		}
		if args.showVersion {
			t.Error("Version must not be shown by default")
		}
		if args.showCredits {
			t.Error("Credits must not be shown by default")
		}
		if v := *args.settings["pid"]; v != "fireworq.pid" {
			t.Errorf("Wrong configuration value for 'pid': %s", v)
		}
		if v := *args.settings["bind"]; v != "0.0.0.0:80" {
			t.Errorf("Wrong configuration value for 'bind': %s", v)
		}
		if v := *args.settings["error_log_level"]; v != "warn" {
			t.Errorf("Wrong configuration value for 'error_log_level': %s", v)
		}
	}
}

func TestInitDefaultConfig(t *testing.T) {
	initDefaultConfig()

	if config.GetDefault("dispatch_user_agent") == "" {
		t.Error("No default value for dispatch_user_agent")
	}
	if config.GetDefault("dispatch_keep_alive") == "" {
		t.Error("No default value for dispatch_keep_alive")
	}
	if config.GetDefault("error_log_level") == "" {
		t.Error("No default value for error_log_level")
	}
	if config.GetDefault("queue_log_level") == "" {
		t.Error("No default value for queue_log_level")
	}
}

var pidRegexp = regexp.MustCompile("^[0-9]+$")

func TestInitProcess(t *testing.T) {
	fname := "test.pid"
	config.Locally("pid", fname, func() {
		initProcess()
		defer os.Remove(fname)

		pid, err := ioutil.ReadFile(fname)
		if err != nil {
			t.Error(err)
		}
		if !pidRegexp.MatchString(strings.TrimSpace(string(pid))) {
			t.Errorf("Wrong format of PID: %s", string(pid))
		}
	})
}

var versionRegexp = regexp.MustCompile("^(?:0|[1-9][0-9]*)(?:[.](?:0|[1-9][0-9]*)){2}(?:[-][0-9A-Za-z-]+(?:[.][0-9A-Za-z-]+)*)?(?:[+][0-9A-Za-z-]+(?:[.][0-9A-Za-z-]+)*)?$")

func TestVersionString(t *testing.T) {
	prerelease := Prerelease
	build := Build
	defer func() {
		Prerelease = prerelease
		Build = build
	}()

	{
		v := strings.SplitN(versionString(" "), " ", 2)
		if len(v) < 2 {
			t.Error("The version string and the application name must be separated")
		}
		if v[0] != "Fireworq" {
			t.Errorf("Wrong application name: %s", v[0])
		}
		if !versionRegexp.MatchString(v[1]) {
			t.Errorf("Wrong version string format: %s", v[1])
		}
	}

	{
		Prerelease = "TEST"
		Build = ""

		v := strings.SplitN(versionString(" "), " ", 2)
		if len(v) < 2 {
			t.Error("The version string and the application name must be separated")
		}
		if v[0] != "Fireworq" {
			t.Errorf("Wrong application name: %s", v[0])
		}
		if !versionRegexp.MatchString(v[1]) {
			t.Errorf("Wrong version string format: %s", v[1])
		}
	}

	{
		Prerelease = ""
		Build = "deadbeafcafe"

		v := strings.SplitN(versionString(" "), " ", 2)
		if len(v) < 2 {
			t.Error("The version string and the application name must be separated")
		}
		if v[0] != "Fireworq" {
			t.Errorf("Wrong application name: %s", v[0])
		}
		if !versionRegexp.MatchString(v[1]) {
			t.Errorf("Wrong version string format: %s", v[1])
		}
	}

	{
		Prerelease = "TEST"
		Build = "deadbeafcafe"

		v := strings.SplitN(versionString(" "), " ", 2)
		if len(v) < 2 {
			t.Error("The version string and the application name must be separated")
		}
		if v[0] != "Fireworq" {
			t.Errorf("Wrong application name: %s", v[0])
		}
		if !versionRegexp.MatchString(v[1]) {
			t.Errorf("Wrong version string format: %s", v[1])
		}
	}

	{
		v := strings.SplitN(versionString("/"), "/", 2)
		if len(v) < 2 {
			t.Error("The version string and the application name must be separated")
		}
		if v[0] != "Fireworq" {
			t.Errorf("Wrong application name: %s", v[0])
		}
		if !versionRegexp.MatchString(v[1]) {
			t.Errorf("Wrong version string format: %s", v[1])
		}
	}

	{
		Prerelease = "TEST"
		Build = ""

		v := strings.SplitN(versionString("/"), "/", 2)
		if len(v) < 2 {
			t.Error("The version string and the application name must be separated")
		}
		if v[0] != "Fireworq" {
			t.Errorf("Wrong application name: %s", v[0])
		}
		if !versionRegexp.MatchString(v[1]) {
			t.Errorf("Wrong version string format: %s", v[1])
		}
	}

	{
		Prerelease = ""
		Build = "deadbeafcafe"

		v := strings.SplitN(versionString("/"), "/", 2)
		if len(v) < 2 {
			t.Error("The version string and the application name must be separated")
		}
		if v[0] != "Fireworq" {
			t.Errorf("Wrong application name: %s", v[0])
		}
		if !versionRegexp.MatchString(v[1]) {
			t.Errorf("Wrong version string format: %s", v[1])
		}
	}

	{
		Prerelease = "TEST"
		Build = "deadbeafcafe"

		v := strings.SplitN(versionString("/"), "/", 2)
		if len(v) < 2 {
			t.Error("The version string and the application name must be separated")
		}
		if v[0] != "Fireworq" {
			t.Errorf("Wrong application name: %s", v[0])
		}
		if !versionRegexp.MatchString(v[1]) {
			t.Errorf("Wrong version string format: %s", v[1])
		}
	}
}

func TestStartServer(t *testing.T) {
	config.Locally("driver", "in-memory", func() {
		config.Locally("bind", "0.0.0.0:0", func() {
			config.Locally("shutdown_timeout", "1", func() {
				stopped := make(chan struct{})
				go func() {
					startServer(os.Stdout)
					stopped <- struct{}{}
				}()
				time.Sleep(1 * time.Second)

				p, err := os.FindProcess(os.Getpid())
				if err != nil {
					t.Error(err)
				}
				if err := p.Signal(syscall.SIGTERM); err != nil {
					t.Error(err)
				}

				<-stopped
			})
		})
	})
}
