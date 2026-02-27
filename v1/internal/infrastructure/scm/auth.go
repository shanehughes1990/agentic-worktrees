package scm

import (
	"context"
	"fmt"
	"strings"
)

type TokenProvider interface {
	AccessToken(ctx context.Context) (string, error)
}

type StaticTokenProvider struct {
	token string
}

func NewStaticTokenProvider(token string) *StaticTokenProvider {
	return &StaticTokenProvider{token: strings.TrimSpace(token)}
}

func (provider *StaticTokenProvider) AccessToken(ctx context.Context) (string, error) {
	_ = ctx
	if provider == nil || strings.TrimSpace(provider.token) == "" {
		return "", fmt.Errorf("scm auth token is not configured")
	}
	return provider.token, nil
}
