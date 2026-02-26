package copilot

import (
	"context"
	"fmt"
	"strings"
)

type Authenticator interface {
	AuthStatus(ctx context.Context) (string, error)
	Authenticate(ctx context.Context) error
	KillOrphanedProcesses(ctx context.Context) (int, error)
}

type AuthService struct {
	authenticator Authenticator
}

func NewAuthService(authenticator Authenticator) *AuthService {
	return &AuthService{authenticator: authenticator}
}

func (service *AuthService) Status(ctx context.Context) (string, error) {
	status, err := service.authenticator.AuthStatus(ctx)
	if err != nil {
		return "", fmt.Errorf("copilot auth status failed: %w", err)
	}
	return strings.TrimSpace(status), nil
}

func (service *AuthService) Authenticate(ctx context.Context) error {
	if err := service.authenticator.Authenticate(ctx); err != nil {
		return fmt.Errorf("copilot authentication failed: %w", err)
	}
	return nil
}

func (service *AuthService) KillOrphanedProcesses(ctx context.Context) (string, error) {
	killedCount, err := service.authenticator.KillOrphanedProcesses(ctx)
	if err != nil {
		return "", fmt.Errorf("copilot process cleanup failed: %w", err)
	}
	if killedCount == 0 {
		return "No orphaned Copilot processes found", nil
	}
	return fmt.Sprintf("Killed %d orphaned Copilot process(es)", killedCount), nil
}
