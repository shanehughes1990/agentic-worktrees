package asynq

import (
	"context"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/shanehughes1990/agentic-worktrees/internal/infrastructure/queue/asynq/tasks"
	"github.com/shanehughes1990/agentic-worktrees/tests/containers"
)

func TestClientEnqueueCopilotDecomposeWithRedisContainer(t *testing.T) {
	redisContainer := containers.StartRedis(t)
	defer redisContainer.Terminate(t)

	cfg, err := NewConfig(redisContainer.URI)
	if err != nil {
		t.Fatalf("unexpected config error: %v", err)
	}

	client := NewClient(cfg)
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Fatalf("unexpected client close error: %v", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := client.EnqueueCopilotDecompose(ctx, tasks.CopilotDecomposePayload{RunID: "run-1", Prompt: "hello"}); err != nil {
		t.Fatalf("unexpected enqueue error: %v", err)
	}

	inspector := asynq.NewInspector(cfg.redisConnOpt)
	defer inspector.Close()

	queueInfo, err := inspector.GetQueueInfo(QueueIngestion)
	if err != nil {
		t.Fatalf("unexpected queue info error: %v", err)
	}
	if queueInfo.Pending < 1 {
		t.Fatalf("expected at least one pending task, got %d", queueInfo.Pending)
	}
}
