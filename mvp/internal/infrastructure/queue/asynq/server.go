package asynq

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

type HandlerRegistration struct {
	TaskType string
	Handler  asynq.Handler
}

type Server struct {
	inner *asynq.Server
}

func NewServer(cfg Config) *Server {
	return &Server{inner: asynq.NewServer(cfg.redisConnOpt, cfg.serverConfig())}
}

func (server *Server) Run(ctx context.Context, registrations []HandlerRegistration) error {
	mux := asynq.NewServeMux()
	for _, registration := range registrations {
		mux.Handle(registration.TaskType, registration.Handler)
	}

	if err := server.inner.Start(mux); err != nil {
		return fmt.Errorf("start asynq server: %w", err)
	}

	<-ctx.Done()
	server.inner.Shutdown()
	return nil
}
