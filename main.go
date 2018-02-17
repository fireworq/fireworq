//go:generate go-assets-builder -p main -o assets.go LICENSE AUTHORS CREDITS
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fireworq/fireworq/config"
	"github.com/fireworq/fireworq/dispatcher"
	"github.com/fireworq/fireworq/jobqueue/logger"
	logwriter "github.com/fireworq/fireworq/log"
	repository "github.com/fireworq/fireworq/repository/factory"
	"github.com/fireworq/fireworq/service"
	"github.com/fireworq/fireworq/web"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	out := os.Stderr

	initDefaultConfig()

	args, err := parseCmdArgs(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	if args.showVersion {
		fmt.Fprintln(out, versionString(" "))
		os.Exit(0)
	}

	if args.showLicense {
		fmt.Println(licenseText)
		os.Exit(0)
	}

	if args.showCredits {
		fmt.Println(creditsText)
		os.Exit(0)
	}

	for _, k := range config.Keys() {
		config.Set(k, *args.settings[k])
	}

	accessLog := initLogging(syscall.SIGUSR1)
	initProcess()
	dispatcher.Init()
	web.Init()

	startServer(accessLog)
}

type cmdArgs struct {
	showVersion bool
	showLicense bool
	showCredits bool
	settings    map[string]*string
}

func parseCmdArgs(args []string) (*cmdArgs, error) {
	out := os.Stderr

	parsed := &cmdArgs{
		settings: make(map[string]*string),
	}

	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(out)
	flags.Usage = func() {
		fmt.Fprint(out, helpText)
		for _, item := range config.Descriptions() {
			fmt.Println("")
			fmt.Fprintf(out, item.Describe(2, 80-4))
		}
	}
	flags.BoolVar(&parsed.showVersion, "v", false, "")
	flags.BoolVar(&parsed.showVersion, "version", false, "")
	flags.BoolVar(&parsed.showLicense, "license", false, "")
	flags.BoolVar(&parsed.showCredits, "credits", false, "")

	for _, k := range config.Keys() {
		p := new(string)
		parsed.settings[k] = p
		name := strings.Replace(k, "_", "-", -1)
		flags.StringVar(p, name, config.Get(k), "")
	}

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	return parsed, nil
}

func initDefaultConfig() {
	config.SetDefault("dispatch_user_agent", versionString("/"))
	config.SetDefault("dispatch_keep_alive", config.Get("keep_alive"))
	if len(os.Getenv("DEBUG")) > 0 {
		config.SetDefault("error_log_level", "debug")
		config.SetDefault("queue_log_level", "debug")
	} else {
		config.SetDefault("error_log_level", "info")
		config.SetDefault("queue_log_level", "info")
	}
}

func initProcess() {
	pid := os.Getpid()
	log.Info().Msgf("PID: %d", pid)

	name := config.Get("pid")
	if name == "" {
		return
	}

	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		log.Error().Msg(err.Error())
		return
	}

	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "%d\n", pid); err != nil {
		log.Error().Msg(err.Error())
	}
}

func initLogging(sig syscall.Signal) (accessLog logwriter.Writer) {
	// Access log

	accessLog = logwriter.New(os.Stdout)

	accessLogFile := config.Get("access_log")
	if len(accessLogFile) > 0 {
		var err error
		accessLog, err = logwriter.OpenFile(accessLogFile)
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}

	// Error log

	errorLog := logwriter.New(zerolog.ConsoleWriter{Out: os.Stderr})

	errorLevel := zerolog.InfoLevel
	errorLevel = logwriter.ParseLevel(config.Get("error_log_level"), errorLevel)
	zerolog.SetGlobalLevel(errorLevel)

	errorLogFile := config.Get("error_log")
	if len(errorLogFile) > 0 {
		var err error
		errorLog, err = logwriter.OpenFile(errorLogFile)
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}
	log.Logger = log.Output(errorLog)

	// Queue log
	logger.Init()

	// Reopening log files (for logrotate)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, sig)
	go func() {
		for {
			s := <-sigc
			log.Info().Msgf("Received signal %q; reopen log files", s)
			if err := accessLog.Reopen(); err != nil {
				log.Error().Msg(err.Error())
			}
			if err := logger.Writer.Reopen(); err != nil {
				log.Error().Msg(err.Error())
			}
			if err := errorLog.Reopen(); err != nil {
				log.Error().Msg(err.Error())
			}
		}
	}()

	return
}

func startServer(accessLogWriter io.Writer) {
	log.Info().Msg("Starting a job dispatcher...")

	repos := repository.NewRepositories()
	service := service.NewService(repos)

	app := &web.Application{
		AccessLogWriter:   accessLogWriter,
		Version:           versionString(" "),
		Service:           service,
		QueueRepository:   repos.Queue,
		RoutingRepository: repos.Routing,
	}
	app.Serve()
}

func versionString(sep string) string {
	var prerelease string
	if Prerelease != "" {
		prerelease = "-" + Prerelease
	}

	var build string
	if Build != "" {
		build = "+" + Build
	}

	return strings.Join([]string{Name, sep, Version, prerelease, build}, "")
}

func mustAssetString(path string) string {
	f, err := Assets.Open(path)
	if err != nil {
		panic(err)
	}

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	return string(buf)
}

var (
	licenseText = mustAssetString("/LICENSE") + `
   Copyright (c) 2017 The Fireworq Authors. All rights reserved.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

` + mustAssetString("/AUTHORS")
	creditsText = mustAssetString("/CREDITS")

	helpText = `Usage: fireworq [options]

  A lightweight, high performance job queue system.

Options:

  --version, -v  Show the version string.
  --license      Show the license text.
  --credits      Show the library dependencies and their licenses.
  --help, -h     Show the help message.
`
)
