package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type CheckResult struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

type PreflightReport struct {
	Healthy bool          `json:"healthy"`
	Checks  []CheckResult `json:"checks"`
}

func runPreflight(ctx context.Context, cfg Config) PreflightReport {
	checks := []CheckResult{
		checkRedis(ctx, cfg.RedisAddr),
		checkQueueName(cfg.QueueName),
		checkBoardPath(cfg.BoardPath),
		checkCheckpointPath(cfg.CheckpointPath),
		checkADKURL(cfg.ADKBoardURL),
	}

	healthy := true
	for _, check := range checks {
		if !check.OK {
			healthy = false
			break
		}
	}

	return PreflightReport{Healthy: healthy, Checks: checks}
}

func checkRedis(ctx context.Context, addr string) CheckResult {
	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()

	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := client.Ping(checkCtx).Err(); err != nil {
		return CheckResult{Name: "redis", OK: false, Detail: err.Error()}
	}
	return CheckResult{Name: "redis", OK: true, Detail: "reachable"}
}

func checkQueueName(queueName string) CheckResult {
	if strings.TrimSpace(queueName) == "" {
		return CheckResult{Name: "asynq_queue", OK: false, Detail: "queue name cannot be empty"}
	}
	return CheckResult{Name: "asynq_queue", OK: true, Detail: queueName}
}

func checkBoardPath(path string) CheckResult {
	if strings.TrimSpace(path) == "" {
		return CheckResult{Name: "board_path", OK: false, Detail: "board path cannot be empty"}
	}
	return CheckResult{Name: "board_path", OK: true, Detail: path}
}

func checkCheckpointPath(path string) CheckResult {
	if strings.TrimSpace(path) == "" {
		return CheckResult{Name: "checkpoint_path", OK: false, Detail: "checkpoint path cannot be empty"}
	}
	return CheckResult{Name: "checkpoint_path", OK: true, Detail: path}
}

func checkADKURL(url string) CheckResult {
	if strings.TrimSpace(url) == "" {
		return CheckResult{Name: "copilot_adk_board_url", OK: false, Detail: "COPILOT_ADK_BOARD_URL is empty"}
	}
	return CheckResult{Name: "copilot_adk_board_url", OK: true, Detail: "configured"}
}

func (r PreflightReport) Error() error {
	if r.Healthy {
		return nil
	}
	failed := make([]string, 0, len(r.Checks))
	for _, check := range r.Checks {
		if !check.OK {
			failed = append(failed, fmt.Sprintf("%s: %s", check.Name, check.Detail))
		}
	}
	return fmt.Errorf("preflight failed: %s", strings.Join(failed, "; "))
}
