package asynq

import (
	"agentic-orchestrator/internal/application/taskengine"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultExecutionJournalTTL = 7 * 24 * time.Hour

type RedisExecutionJournal struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisExecutionJournal(client *redis.Client, ttl time.Duration) *RedisExecutionJournal {
	if ttl <= 0 {
		ttl = defaultExecutionJournalTTL
	}
	return &RedisExecutionJournal{client: client, ttl: ttl}
}

func NewRedisExecutionJournalFromConfig(config Config, ttl time.Duration) *RedisExecutionJournal {
	normalized := config.normalized()
	client := redis.NewClient(&redis.Options{
		Addr:     normalized.RedisAddress,
		Password: normalized.RedisPassword,
		DB:       normalized.RedisDatabase,
	})
	return NewRedisExecutionJournal(client, ttl)
}

func (journal *RedisExecutionJournal) Upsert(ctx context.Context, record taskengine.ExecutionRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("execution journal: marshal: %w", err)
	}
	return journal.client.Set(ctx, executionKey(record.RunID, record.TaskID, record.JobID), payload, journal.ttl).Err()
}

func (journal *RedisExecutionJournal) Load(ctx context.Context, runID string, taskID string, jobID string) (*taskengine.ExecutionRecord, error) {
	if strings.TrimSpace(runID) == "" || strings.TrimSpace(taskID) == "" || strings.TrimSpace(jobID) == "" {
		return nil, fmt.Errorf("execution journal: run_id, task_id, job_id are required")
	}
	payload, err := journal.client.Get(ctx, executionKey(runID, taskID, jobID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("execution journal: load: %w", err)
	}
	var record taskengine.ExecutionRecord
	if err := json.Unmarshal(payload, &record); err != nil {
		return nil, fmt.Errorf("execution journal: unmarshal: %w", err)
	}
	if err := record.Validate(); err != nil {
		return nil, err
	}
	return &record, nil
}

func executionKey(runID string, taskID string, jobID string) string {
	return fmt.Sprintf("execution:%s:%s:%s", strings.TrimSpace(runID), strings.TrimSpace(taskID), strings.TrimSpace(jobID))
}
