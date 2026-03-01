package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	applicationworker "agentic-orchestrator/internal/application/worker"

	"github.com/jackc/pgx/v5"
	"gorm.io/gorm"
)

const (
	heartbeatRequestChannel  = "worker_heartbeat_request"
	heartbeatResponseChannel = "worker_heartbeat_response"
)

type PGNotifyTransport struct {
	db  *gorm.DB
	dsn string
}

func NewPGNotifyTransport(db *gorm.DB, dsn string) (*PGNotifyTransport, error) {
	if db == nil {
		return nil, fmt.Errorf("pg notify transport db is required")
	}
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("pg notify transport dsn is required")
	}
	return &PGNotifyTransport{db: db, dsn: strings.TrimSpace(dsn)}, nil
}

func (transport *PGNotifyTransport) PublishRequest(ctx context.Context, request applicationworker.HeartbeatRequest) error {
	if transport == nil || transport.db == nil {
		return fmt.Errorf("pg notify transport is not initialized")
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal heartbeat request: %w", err)
	}
	if err := transport.db.WithContext(ctx).Exec("SELECT pg_notify(?, ?)", heartbeatRequestChannel, string(payload)).Error; err != nil {
		return fmt.Errorf("publish heartbeat request: %w", err)
	}
	return nil
}

func (transport *PGNotifyTransport) PublishResponse(ctx context.Context, response applicationworker.HeartbeatResponse) error {
	if transport == nil || transport.db == nil {
		return fmt.Errorf("pg notify transport is not initialized")
	}
	payload, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal heartbeat response: %w", err)
	}
	if err := transport.db.WithContext(ctx).Exec("SELECT pg_notify(?, ?)", heartbeatResponseChannel, string(payload)).Error; err != nil {
		return fmt.Errorf("publish heartbeat response: %w", err)
	}
	return nil
}

func (transport *PGNotifyTransport) ListenRequests(ctx context.Context, handler func(applicationworker.HeartbeatRequest) error) error {
	return transport.listen(ctx, heartbeatRequestChannel, func(payload []byte) error {
		request := applicationworker.HeartbeatRequest{}
		if err := json.Unmarshal(payload, &request); err != nil {
			return fmt.Errorf("decode heartbeat request: %w", err)
		}
		return handler(request)
	})
}

func (transport *PGNotifyTransport) ListenResponses(ctx context.Context, handler func(applicationworker.HeartbeatResponse) error) error {
	return transport.listen(ctx, heartbeatResponseChannel, func(payload []byte) error {
		response := applicationworker.HeartbeatResponse{}
		if err := json.Unmarshal(payload, &response); err != nil {
			return fmt.Errorf("decode heartbeat response: %w", err)
		}
		return handler(response)
	})
}

func (transport *PGNotifyTransport) listen(ctx context.Context, channel string, decode func([]byte) error) error {
	if transport == nil || strings.TrimSpace(transport.dsn) == "" {
		return fmt.Errorf("pg notify transport dsn is not configured")
	}
	connection, err := pgx.Connect(ctx, transport.dsn)
	if err != nil {
		return fmt.Errorf("connect pgx listener: %w", err)
	}
	defer connection.Close(ctx)
	if _, err := connection.Exec(ctx, fmt.Sprintf("LISTEN %s", channel)); err != nil {
		return fmt.Errorf("listen on %s: %w", channel, err)
	}
	for {
		notification, waitErr := connection.WaitForNotification(ctx)
		if waitErr != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("wait notification: %w", waitErr)
		}
		if decodeErr := decode([]byte(notification.Payload)); decodeErr != nil {
			return decodeErr
		}
	}
}

var _ applicationworker.HeartbeatTransport = (*PGNotifyTransport)(nil)
