package asynq

import (
	"context"

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

	runErr := make(chan error, 1)
	go func() {
		runErr <- server.inner.Run(mux)
	}()

	select {
	case <-ctx.Done():
		server.inner.Shutdown()
		return <-runErr
	case err := <-runErr:
		return err
	}
}
