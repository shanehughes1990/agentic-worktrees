package copilot

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type Authenticator interface {
	AuthStatus(ctx context.Context) (string, error)
	Authenticate(ctx context.Context) error
	KillOrphanedProcesses(ctx context.Context) (int, error)
}

type AuthService struct {
	authenticator Authenticator
	logger        *logrus.Logger
}

func NewAuthService(authenticator Authenticator, loggers ...*logrus.Logger) *AuthService {
	var logger *logrus.Logger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &AuthService{authenticator: authenticator, logger: logger}
}

func (service *AuthService) Status(ctx context.Context) (string, error) {
	entry := service.entry().WithField("event", "copilot.auth.status")
	status, err := service.authenticator.AuthStatus(ctx)
	if err != nil {
		entry.WithError(err).Error("copilot auth status failed")
		return "", fmt.Errorf("copilot auth status failed: %w", err)
	}
	entry.Info("copilot auth status succeeded")
	return strings.TrimSpace(status), nil
}

func (service *AuthService) Authenticate(ctx context.Context) error {
	entry := service.entry().WithField("event", "copilot.auth.authenticate")
	if err := service.authenticator.Authenticate(ctx); err != nil {
		entry.WithError(err).Error("copilot authentication failed")
		return fmt.Errorf("copilot authentication failed: %w", err)
	}
	entry.Info("copilot authentication succeeded")
	return nil
}

func (service *AuthService) KillOrphanedProcesses(ctx context.Context) (string, error) {
	entry := service.entry().WithField("event", "copilot.auth.kill_orphaned")
	killedCount, err := service.authenticator.KillOrphanedProcesses(ctx)
	if err != nil {
		entry.WithError(err).Error("copilot process cleanup failed")
		return "", fmt.Errorf("copilot process cleanup failed: %w", err)
	}
	entry.WithField("killed_count", killedCount).Info("copilot orphan process cleanup completed")
	if killedCount == 0 {
		return "No orphaned Copilot processes found", nil
	}
	return fmt.Sprintf("Killed %d orphaned Copilot process(es)", killedCount), nil
}

func (service *AuthService) entry() *logrus.Entry {
	if service == nil || service.logger == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	return logrus.NewEntry(service.logger)
}
