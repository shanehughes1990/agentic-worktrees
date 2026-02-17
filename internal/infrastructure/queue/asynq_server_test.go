package queue

import "testing"

func TestResolveRedisClientOptAcceptsHostPort(t *testing.T) {
	if _, err := ResolveRedisClientOpt("127.0.0.1:6379"); err != nil {
		t.Fatalf("expected host:port to be accepted, got %v", err)
	}
}

func TestResolveRedisClientOptAcceptsRedisURI(t *testing.T) {
	if _, err := ResolveRedisClientOpt("redis://127.0.0.1:6379"); err != nil {
		t.Fatalf("expected redis uri to be accepted, got %v", err)
	}
}

func TestResolveRedisClientOptRejectsEmpty(t *testing.T) {
	if _, err := ResolveRedisClientOpt("   "); err == nil {
		t.Fatalf("expected empty redis address error")
	}
}

func TestNewAsynqServerValidation(t *testing.T) {
	if _, err := NewAsynqServer("127.0.0.1:6379", "", 1); err == nil {
		t.Fatalf("expected queue name validation error")
	}
	if _, err := NewAsynqServer("127.0.0.1:6379", "default", 0); err == nil {
		t.Fatalf("expected concurrency validation error")
	}
}
