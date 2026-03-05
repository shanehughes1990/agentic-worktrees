package realtime

import (
	domainrealtime "agentic-orchestrator/internal/domain/realtime"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"gorm.io/gorm"
)

var watcherTopicPattern = regexp.MustCompile(`^[a-z0-9_]+$`)

type PGTableChangeWatcher struct {
	db  *gorm.DB
	dsn string
}

func NewPGTableChangeWatcher(db *gorm.DB, dsn string) (*PGTableChangeWatcher, error) {
	if db == nil {
		return nil, fmt.Errorf("pg table change watcher db is required")
	}
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("pg table change watcher dsn is required")
	}
	return &PGTableChangeWatcher{db: db, dsn: strings.TrimSpace(dsn)}, nil
}

func (watcher *PGTableChangeWatcher) Publish(ctx context.Context, event domainrealtime.TableChangeEvent) error {
	if watcher == nil || watcher.db == nil {
		return fmt.Errorf("pg table change watcher is not initialized")
	}
	if err := event.Validate(); err != nil {
		return err
	}
	topic := strings.TrimSpace(event.Topic)
	if !watcherTopicPattern.MatchString(topic) {
		return fmt.Errorf("pg table change watcher topic %q must match %s", topic, watcherTopicPattern.String())
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal table change event: %w", err)
	}
	if err := watcher.db.WithContext(ctx).Exec("SELECT pg_notify(?, ?)", topic, string(payload)).Error; err != nil {
		return fmt.Errorf("publish table change event: %w", err)
	}
	return nil
}

func (watcher *PGTableChangeWatcher) Watch(ctx context.Context, topic string, handler func(domainrealtime.TableChangeEvent) error) error {
	if watcher == nil || strings.TrimSpace(watcher.dsn) == "" {
		return fmt.Errorf("pg table change watcher dsn is not configured")
	}
	normalizedTopic := strings.TrimSpace(topic)
	if !watcherTopicPattern.MatchString(normalizedTopic) {
		return fmt.Errorf("pg table change watcher topic %q must match %s", normalizedTopic, watcherTopicPattern.String())
	}
	connection, err := pgx.Connect(ctx, watcher.dsn)
	if err != nil {
		return fmt.Errorf("connect pgx watcher: %w", err)
	}
	defer connection.Close(ctx)
	if _, err := connection.Exec(ctx, fmt.Sprintf("LISTEN %s", normalizedTopic)); err != nil {
		return fmt.Errorf("listen on table-change topic %s: %w", normalizedTopic, err)
	}
	for {
		notification, waitErr := connection.WaitForNotification(ctx)
		if waitErr != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("wait table-change notification: %w", waitErr)
		}
		event := domainrealtime.TableChangeEvent{}
		if err := json.Unmarshal([]byte(notification.Payload), &event); err != nil {
			return fmt.Errorf("decode table-change notification payload: %w", err)
		}
		if err := event.Validate(); err != nil {
			return err
		}
		if err := handler(event); err != nil {
			return err
		}
	}
}

var _ domainrealtime.TableChangeWatcher = (*PGTableChangeWatcher)(nil)
