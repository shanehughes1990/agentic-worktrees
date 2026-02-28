package asynq

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultCheckpointTTL = 24 * time.Hour

// RedisCheckpointStore persists taskengine.Checkpoint values in Redis so that
// retried jobs can resume from the last completed step.
type RedisCheckpointStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCheckpointStore returns a RedisCheckpointStore backed by the provided client.
// A non-positive ttl falls back to defaultCheckpointTTL.
func NewRedisCheckpointStore(client *redis.Client, ttl time.Duration) *RedisCheckpointStore {
	if ttl <= 0 {
		ttl = defaultCheckpointTTL
	}
	return &RedisCheckpointStore{client: client, ttl: ttl}
}

// NewRedisCheckpointStoreFromConfig constructs a RedisCheckpointStore from the
// same Config used by the asynq Platform so callers share Redis connection parameters.
func NewRedisCheckpointStoreFromConfig(config Config, ttl time.Duration) *RedisCheckpointStore {
	normalized := config.normalized()
	client := redis.NewClient(&redis.Options{
		Addr:     normalized.RedisAddress,
		Password: normalized.RedisPassword,
		DB:       normalized.RedisDatabase,
	})
	return NewRedisCheckpointStore(client, ttl)
}

func (store *RedisCheckpointStore) Save(ctx context.Context, idempotencyKey string, checkpoint taskengine.Checkpoint) error {
	if strings.TrimSpace(idempotencyKey) == "" {
		return fmt.Errorf("checkpoint store: idempotency_key is required")
	}
	data, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("checkpoint store: marshal: %w", err)
	}
	return store.client.Set(ctx, checkpointKey(idempotencyKey), data, store.ttl).Err()
}

func (store *RedisCheckpointStore) Load(ctx context.Context, idempotencyKey string) (*taskengine.Checkpoint, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return nil, fmt.Errorf("checkpoint store: idempotency_key is required")
	}
	data, err := store.client.Get(ctx, checkpointKey(idempotencyKey)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("checkpoint store: load: %w", err)
	}
	var checkpoint taskengine.Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("checkpoint store: unmarshal: %w", err)
	}
	return &checkpoint, nil
}

func checkpointKey(idempotencyKey string) string {
	return "checkpoint:" + idempotencyKey
}
