package web

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fireworq/fireworq/config"

	"github.com/gorilla/mux"
	serverstarter "github.com/lestrrat/go-server-starter/listener"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

func graceful(server *http.Server, shutdownTimeout time.Duration) (time.Duration, error) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	sig := <-sigc
	log.Info().Msgf(
		"Received signal %q; shutting down gracefully in %s ...",
		sig,
		shutdownTimeout,
	)
	deadline := time.Now().Add(shutdownTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	errc := make(chan error)
	go func() { errc <- server.Shutdown(ctx) }()

	select {
	case sig := <-sigc:
		log.Info().Msgf("Received second signal %q; shutdown now", sig)
		cancel()
		return 0, server.Close()
	case err := <-errc:
		return time.Until(deadline), err
	}
}

type server struct {
	addrs       []net.Addr
	makeHandler func(h http.Handler) http.Handler
	mux         *mux.Router
}

func newServer(out io.Writer) *server {
	tag := config.Get("access_log_tag")
	logger := zerolog.New(out).With().
		Timestamp().
		Logger()
	accessLog := hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("tag", tag).
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Int("status", status).
			Dur("duration", duration).
			Msg("")
	})
	remoteAddr := hlog.RemoteAddrHandler("remote_addr")
	ua := hlog.UserAgentHandler("user_agent")

	s := &server{
		makeHandler: func(h http.Handler) http.Handler {
			return hlog.NewHandler(logger)(accessLog(remoteAddr(ua(h))))
		},
		mux: mux.NewRouter(),
	}
	return s
}

func (s *server) start() (*http.Server, error) {
	server := &http.Server{Handler: s.mux}

	listeners, err := serverstarter.ListenAll()
	if err == serverstarter.ErrNoListeningTarget {
		log.Info().Msg("Starting a server ...")

		ln, err := net.Listen("tcp", config.Get("bind"))
		if err != nil {
			return nil, err
		}

		listeners = []net.Listener{ln}
	} else if err != nil {
		return nil, err
	} else {
		log.Info().Msg("Starting a server under start_server ...")
	}

	addrs := make([]net.Addr, 0)
	for _, ln := range listeners {
		ln := ln
		addr := ln.Addr()
		addrs = append(addrs, addr)
		log.Info().Msgf("Listening on %s", addr.String())

		go func() {
			err := server.Serve(ln)
			log.Info().Msg(err.Error())
		}()
	}
	s.addrs = addrs

	return server, nil
}

func (s *server) startGracefully(shutdownTimeout time.Duration) (time.Duration, error) {
	server, err := s.start()
	if err != nil {
		return 0, err
	}

	return graceful(server, shutdownTimeout)
}

func (s *server) handle(pattern string, h func(http.ResponseWriter, *http.Request) error) {
	s.mux.Handle(pattern, s.makeHandler(handler(h)))
}
