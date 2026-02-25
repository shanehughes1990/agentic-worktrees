package copilot

import (
	"context"
	"fmt"
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
