package scm

import (
	"context"
	"testing"
)

func TestStaticTokenProviderRejectsEmptyToken(t *testing.T) {
	provider := NewStaticTokenProvider(" ")
	_, err := provider.AccessToken(context.Background())
	if err == nil {
		t.Fatalf("expected error for empty token")
	}
}
