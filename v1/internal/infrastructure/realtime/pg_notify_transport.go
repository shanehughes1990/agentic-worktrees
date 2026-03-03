package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	domainrealtime "agentic-orchestrator/internal/domain/realtime"

	"github.com/jackc/pgx/v5"
	"gorm.io/gorm"
)

const (
	heartbeatRequestChannel       = "worker_heartbeat_request"
	heartbeatResponseChannel      = "worker_heartbeat_response"
	registrationSubmissionChannel = "worker_registration_submission"
	registrationDecisionChannel   = "worker_registration_decision"
	invalidationIntentChannel     = "worker_invalidation_intent"
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

func (transport *PGNotifyTransport) PublishRequest(ctx context.Context, request domainrealtime.HeartbeatRequest) error {
	if transport == nil || transport.db == nil {
		return fmt.Errorf("pg notify transport is not initialized")
	}
	if err := request.Validate(); err != nil {
		return err
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

func (transport *PGNotifyTransport) PublishResponse(ctx context.Context, response domainrealtime.HeartbeatResponse) error {
	if transport == nil || transport.db == nil {
		return fmt.Errorf("pg notify transport is not initialized")
	}
	if err := response.Validate(); err != nil {
		return err
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

func (transport *PGNotifyTransport) ListenRequests(ctx context.Context, handler func(domainrealtime.HeartbeatRequest) error) error {
	return transport.listen(ctx, heartbeatRequestChannel, func(payload []byte) error {
		request := domainrealtime.HeartbeatRequest{}
		if err := json.Unmarshal(payload, &request); err != nil {
			return fmt.Errorf("decode heartbeat request: %w", err)
		}
		if err := request.Validate(); err != nil {
			return err
		}
		return handler(request)
	})
}

func (transport *PGNotifyTransport) ListenResponses(ctx context.Context, handler func(domainrealtime.HeartbeatResponse) error) error {
	return transport.listen(ctx, heartbeatResponseChannel, func(payload []byte) error {
		response := domainrealtime.HeartbeatResponse{}
		if err := json.Unmarshal(payload, &response); err != nil {
			return fmt.Errorf("decode heartbeat response: %w", err)
		}
		if err := response.Validate(); err != nil {
			return err
		}
		return handler(response)
	})
}

func (transport *PGNotifyTransport) PublishRegistrationSubmission(ctx context.Context, event domainrealtime.RegistrationSubmissionEvent) error {
	if transport == nil || transport.db == nil {
		return fmt.Errorf("pg notify transport is not initialized")
	}
	if err := event.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal registration submission: %w", err)
	}
	if err := transport.db.WithContext(ctx).Exec("SELECT pg_notify(?, ?)", registrationSubmissionChannel, string(payload)).Error; err != nil {
		return fmt.Errorf("publish registration submission: %w", err)
	}
	return nil
}

func (transport *PGNotifyTransport) PublishRegistrationDecision(ctx context.Context, event domainrealtime.RegistrationDecisionEvent) error {
	if transport == nil || transport.db == nil {
		return fmt.Errorf("pg notify transport is not initialized")
	}
	if err := event.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal registration decision: %w", err)
	}
	if err := transport.db.WithContext(ctx).Exec("SELECT pg_notify(?, ?)", registrationDecisionChannel, string(payload)).Error; err != nil {
		return fmt.Errorf("publish registration decision: %w", err)
	}
	return nil
}

func (transport *PGNotifyTransport) ListenRegistrationSubmissions(ctx context.Context, handler func(domainrealtime.RegistrationSubmissionEvent) error) error {
	return transport.listen(ctx, registrationSubmissionChannel, func(payload []byte) error {
		event := domainrealtime.RegistrationSubmissionEvent{}
		if err := json.Unmarshal(payload, &event); err != nil {
			return fmt.Errorf("decode registration submission: %w", err)
		}
		if err := event.Validate(); err != nil {
			return err
		}
		return handler(event)
	})
}

func (transport *PGNotifyTransport) ListenRegistrationDecisions(ctx context.Context, handler func(domainrealtime.RegistrationDecisionEvent) error) error {
	return transport.listen(ctx, registrationDecisionChannel, func(payload []byte) error {
		event := domainrealtime.RegistrationDecisionEvent{}
		if err := json.Unmarshal(payload, &event); err != nil {
			return fmt.Errorf("decode registration decision: %w", err)
		}
		if err := event.Validate(); err != nil {
			return err
		}
		return handler(event)
	})
}

func (transport *PGNotifyTransport) PublishInvalidationIntent(ctx context.Context, intent domainrealtime.InvalidationIntent) error {
	if transport == nil || transport.db == nil {
		return fmt.Errorf("pg notify transport is not initialized")
	}
	if err := intent.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(intent)
	if err != nil {
		return fmt.Errorf("marshal invalidation intent: %w", err)
	}
	if err := transport.db.WithContext(ctx).Exec("SELECT pg_notify(?, ?)", invalidationIntentChannel, string(payload)).Error; err != nil {
		return fmt.Errorf("publish invalidation intent: %w", err)
	}
	return nil
}

func (transport *PGNotifyTransport) ListenInvalidationIntents(ctx context.Context, handler func(domainrealtime.InvalidationIntent) error) error {
	return transport.listen(ctx, invalidationIntentChannel, func(payload []byte) error {
		intent := domainrealtime.InvalidationIntent{}
		if err := json.Unmarshal(payload, &intent); err != nil {
			return fmt.Errorf("decode invalidation intent: %w", err)
		}
		if err := intent.Validate(); err != nil {
			return err
		}
		return handler(intent)
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

var _ domainrealtime.WorkerLifecycleTransport = (*PGNotifyTransport)(nil)
