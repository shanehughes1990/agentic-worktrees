package copilot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	sdk "github.com/github/copilot-sdk/go"
	"github.com/sirupsen/logrus"
)

type Authenticator struct {
	config ClientConfig
	logger *logrus.Logger
}

func NewAuthenticator(config ClientConfig, logger *logrus.Logger) *Authenticator {
	return &Authenticator{config: config.Normalized(), logger: logger}
}

func (authenticator *Authenticator) AuthStatus(ctx context.Context) (string, error) {
	status, err := authenticator.fetchAuthStatus(ctx)
	if err != nil {
		authenticator.entry().WithError(err).Error("copilot sdk auth status check failed")
		return "", err
	}

	formatted := formatSDKAuthStatus(status)
	authenticator.entry().WithFields(logrus.Fields{
		"event":            "copilot.auth.status",
		"is_authenticated": status.IsAuthenticated,
		"auth_type":        safeStringPtr(status.AuthType),
		"login":            safeStringPtr(status.Login),
		"host":             safeStringPtr(status.Host),
		"status_message":   safeStringPtr(status.StatusMessage),
	}).Info("copilot sdk auth status resolved")

	return formatted, nil
}

func (authenticator *Authenticator) Authenticate(ctx context.Context) error {
	status, err := authenticator.fetchAuthStatus(ctx)
	if err != nil {
		return err
	}
	if status.IsAuthenticated {
		authenticator.entry().WithFields(logrus.Fields{"event": "copilot.auth.validate", "login": safeStringPtr(status.Login)}).Info("copilot non-interactive auth already valid")
		return nil
	}

	message := strings.TrimSpace(formatSDKAuthStatus(status))
	if message == "" {
		message = "not authenticated"
	}
	return fmt.Errorf("non-interactive auth is required. current_status=%s. configure one of COPILOT_GITHUB_TOKEN, GH_TOKEN, or GITHUB_TOKEN (or ensure stored CLI login is available without interactive prompts)", message)
}

func (authenticator *Authenticator) KillOrphanedProcesses(ctx context.Context) (int, error) {
	pids, err := authenticator.findCopilotProcessPIDs(ctx)
	if err != nil {
		return 0, err
	}
	if len(pids) == 0 {
		return 0, nil
	}

	currentPID := os.Getpid()
	killedCount := 0
	for _, pid := range pids {
		if pid <= 1 || pid == currentPID {
			continue
		}
		process, findErr := os.FindProcess(pid)
		if findErr != nil {
			continue
		}
		if signalErr := process.Signal(os.Interrupt); signalErr != nil {
			continue
		}
		killedCount++
	}

	if killedCount == 0 {
		return 0, nil
	}

	time.Sleep(350 * time.Millisecond)
	remainingPIDs, remainingErr := authenticator.findCopilotProcessPIDs(ctx)
	if remainingErr == nil {
		for _, pid := range remainingPIDs {
			if pid <= 1 || pid == currentPID {
				continue
			}
			process, findErr := os.FindProcess(pid)
			if findErr != nil {
				continue
			}
			_ = process.Kill()
		}
	}

	return killedCount, nil
}

func (authenticator *Authenticator) fetchAuthStatus(ctx context.Context) (*sdk.GetAuthStatusResponse, error) {
	checkCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	compatCommand, compatErr := runSDKCLICompatibilityCheck(checkCtx, authenticator.config.CLIPath)
	if compatErr != nil {
		return nil, fmt.Errorf("copilot sdk preflight compatibility failed (command=%q): %w", compatCommand, compatErr)
	}

	client := sdk.NewClient(&sdk.ClientOptions{
		GitHubToken: authenticator.config.GitHubToken,
		CLIPath:     authenticator.config.CLIPath,
		CLIUrl:      authenticator.config.CLIURL,
		LogLevel:    authenticator.config.LogLevel,
	})
	if err := client.Start(checkCtx); err != nil {
		return nil, fmt.Errorf("start copilot sdk client for auth check: %s: %w", explainClientStartError(err), err)
	}
	defer client.Stop()

	status, err := client.GetAuthStatus(checkCtx)
	if err != nil {
		return nil, fmt.Errorf("copilot sdk auth status request failed: %w", err)
	}
	if status == nil {
		return nil, fmt.Errorf("copilot sdk auth status unavailable")
	}
	return status, nil
}

func (authenticator *Authenticator) entry() *logrus.Entry {
	if authenticator.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(authenticator.logger)
}

func (authenticator *Authenticator) findCopilotProcessPIDs(ctx context.Context) ([]int, error) {
	commandContext, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	command := exec.CommandContext(commandContext, "pgrep", "-f", "copilot.*(--headless|--stdio|--allow-all-tools)")
	output, err := command.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("find copilot processes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	pids := make([]int, 0, len(lines))
	for _, line := range lines {
		cleanLine := strings.TrimSpace(line)
		if cleanLine == "" {
			continue
		}
		pid, parseErr := strconv.Atoi(cleanLine)
		if parseErr != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}
