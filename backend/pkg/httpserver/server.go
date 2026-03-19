package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 120 * time.Second
	defaultShutdownTime = 10 * time.Second
)

type Server struct {
	server       *http.Server
	shutdownTime time.Duration
}

func New(handler http.Handler, port int, opts ...Option) *Server {
	s := &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      handler,
			ReadTimeout:  defaultReadTimeout,
			WriteTimeout: defaultWriteTimeout,
			IdleTimeout:  defaultIdleTimeout,
		},
		shutdownTime: defaultShutdownTime,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.shutdownTime)
	defer cancel()

	return s.server.Shutdown(ctx)
}

type Option func(*Server)

func WithReadTimeout(t time.Duration) Option {
	return func(s *Server) {
		s.server.ReadTimeout = t
	}
}

func WithWriteTimeout(t time.Duration) Option {
	return func(s *Server) {
		s.server.WriteTimeout = t
	}
}

func WithShutdownTimeout(t time.Duration) Option {
	return func(s *Server) {
		s.shutdownTime = t
	}
}
