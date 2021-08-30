//go:build linux || darwin

package main

import (
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/fireworq/fireworq/config"
	logwriter "github.com/fireworq/fireworq/log"
)

var (
	testAccessLog = "test-access.log"
	testErrorLog  = "test-error.log"
	testSig       = syscall.SIGUSR2

	testAccessLogWriter logwriter.Writer
)

func initTestMain() {
	config.Locally("access_log", testAccessLog, func() {
		config.Locally("error_log", testErrorLog, func() {
			testAccessLogWriter = initLogging(testSig)
		})
	})
}

func cleanTestMain() {
	os.Remove(testAccessLog)
	os.Remove(testErrorLog)
}

func TestInitLogging(t *testing.T) {
	config.Locally("access_log", testAccessLog, func() {
		config.Locally("error_log", testErrorLog, func() {
			defer func() {
				os.Remove(testAccessLog)
				os.Remove(testErrorLog)
			}()

			testAccessLogWriter.Write([]byte("Some access log\n"))
			log.Warn().Msg("Some warning")

			var accessLogContent string

			{
				buf, err := ioutil.ReadFile(testAccessLog)
				if err != nil {
					t.Error(err)
				}
				accessLogContent = string(buf)
				if accessLogContent == "" {
					t.Error("A log entry must be written to a file")
				}
			}
			{
				buf, err := ioutil.ReadFile(testErrorLog)
				if err != nil {
					t.Error(err)
				}
				if string(buf) == "" {
					t.Error("A log entry must be written to a file")
				}
			}

			movedAccessLog := "test-moved-access.log"
			movedErrorLog := "test-moved-error.log"
			os.Rename(testAccessLog, movedAccessLog)
			os.Rename(testErrorLog, movedErrorLog)
			defer func() {
				os.Remove(movedAccessLog)
				os.Remove(movedErrorLog)
			}()

			p, err := os.FindProcess(os.Getpid())
			if err != nil {
				t.Error(err)
			}
			if err := p.Signal(testSig); err != nil {
				t.Error(err)
			}
			time.Sleep(300 * time.Millisecond)

			otherAccessContent := "Some other access log\n"
			testAccessLogWriter.Write([]byte(otherAccessContent))
			otherErrorContent := "Some other warning"
			log.Warn().Msg(otherErrorContent)

			{
				buf, err := ioutil.ReadFile(testAccessLog)
				if err != nil {
					t.Error(err)
				}
				if !strings.Contains(string(buf), otherAccessContent) {
					t.Error("A log entry must be written to a file")
				}
			}
			{
				buf, err := ioutil.ReadFile(testErrorLog)
				if err != nil {
					t.Error(err)
				}
				if !strings.Contains(string(buf), otherErrorContent) {
					t.Error("A log entry must be written to a file")
				}
			}

			{
				buf, err := ioutil.ReadFile(movedAccessLog)
				if err != nil {
					t.Error(err)
				}
				if string(buf) != accessLogContent {
					t.Error("Moved log file must be kept untouched")
				}
			}
			{
				buf, err := ioutil.ReadFile(movedErrorLog)
				if err != nil {
					t.Error(err)
				}
				if strings.Contains(string(buf), otherErrorContent) {
					t.Error("Moved log file must be kept untouched")
				}
			}
		})
	})
}
