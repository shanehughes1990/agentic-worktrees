package copilot

import "testing"

func TestClientConfigAuthCommandDefaults(t *testing.T) {
	cfg := (ClientConfig{}).Normalized()
	if cfg.AuthStatusCommand != "copilot auth status" {
		t.Fatalf("expected non-interactive auth status default, got %q", cfg.AuthStatusCommand)
	}
	if cfg.AuthLoginCommand != "copilot auth login" {
		t.Fatalf("expected non-interactive auth login default, got %q", cfg.AuthLoginCommand)
	}
}
