package app

import (
	"context"
	"fmt"
	"os/exec"
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

func RunPreflight(ctx context.Context, cfg Config) PreflightReport {
	checks := []CheckResult{
		checkGoBinary(ctx),
		checkGitBinary(ctx),
		checkRedis(ctx, cfg.RedisAddr),
		checkModel(cfg.ModelDefault),
		checkAuditPath(cfg.AuditLogPath),
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

func (r PreflightReport) Error() error {
	if r.Healthy {
		return nil
	}
	parts := make([]string, 0, len(r.Checks))
	for _, check := range r.Checks {
		if !check.OK {
			parts = append(parts, fmt.Sprintf("%s: %s", check.Name, check.Detail))
		}
	}
	return fmt.Errorf("preflight failed: %s", strings.Join(parts, "; "))
}

func checkGoBinary(ctx context.Context) CheckResult {
	_, err := exec.LookPath("go")
	if err != nil {
		return CheckResult{Name: "go", OK: false, Detail: "go binary not found"}
	}
	return CheckResult{Name: "go", OK: true, Detail: "available"}
}

func checkGitBinary(ctx context.Context) CheckResult {
	_, err := exec.LookPath("git")
	if err != nil {
		return CheckResult{Name: "git", OK: false, Detail: "git binary not found"}
	}
	return CheckResult{Name: "git", OK: true, Detail: "available"}
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

func checkModel(model string) CheckResult {
	if strings.TrimSpace(model) == "" {
		return CheckResult{Name: "model-default", OK: false, Detail: "MODEL_DEFAULT is empty"}
	}
	return CheckResult{Name: "model-default", OK: true, Detail: model}
}

func checkAuditPath(path string) CheckResult {
	if strings.TrimSpace(path) == "" {
		return CheckResult{Name: "audit-log", OK: false, Detail: "AUDIT_LOG_PATH is empty"}
	}
	return CheckResult{Name: "audit-log", OK: true, Detail: path}
}
