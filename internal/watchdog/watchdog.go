package watchdog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/shanehughes1990/agentic-worktrees/internal/app"
	"github.com/shanehughes1990/agentic-worktrees/internal/queue"
)

func Run(ctx context.Context, redisAddr string, interval time.Duration, audit *app.AuditSink, metrics *queue.Metrics) {
	if interval < time.Second {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			client := redis.NewClient(&redis.Options{Addr: redisAddr})
			err := client.Ping(pingCtx).Err()
			_ = client.Close()
			cancel()

			snapshot := metrics.Snapshot()
			data, _ := json.Marshal(map[string]any{
				"redis_ok":  err == nil,
				"in_flight": snapshot.InFlight,
				"started":   snapshot.Started,
				"completed": snapshot.Completed,
				"failed":    snapshot.Failed,
			})
			_ = audit.Write(ctx, app.AuditEvent{Type: "watchdog.tick", Data: data})
		}
	}
}
