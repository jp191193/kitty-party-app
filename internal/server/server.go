// Package server wraps the standard net/http Server to support
// graceful shutdown on OS signals.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// Server wraps an *http.Server and provides graceful shutdown.
type Server struct {
	httpServer *http.Server
	log        *zap.Logger
}

// New creates a Server instance.
func New(handler http.Handler, port string, log *zap.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%s", port),
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		log: log,
	}
}

// Run starts the server and blocks until an OS signal is received,
// then performs a graceful shutdown with a 10-second timeout.
func (s *Server) Run() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s.log.Info("server starting", zap.String("addr", s.httpServer.Addr))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Fatal("server error", zap.Error(err))
		}
	}()

	<-quit
	s.log.Info("shutting down server gracefully…")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Error("server forced to shutdown", zap.Error(err))
	}

	s.log.Info("server stopped")
}
