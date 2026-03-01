package postgres

import "testing"

func TestIsPostgresScheme(t *testing.T) {
	if !isPostgresScheme("postgres") {
		t.Fatalf("expected postgres scheme to be supported")
	}
	if !isPostgresScheme("postgresql") {
		t.Fatalf("expected postgresql scheme to be supported")
	}
	if isPostgresScheme("mysql") {
		t.Fatalf("expected mysql scheme to be rejected")
	}
}

func TestConfigNormalizedAppliesDefaults(t *testing.T) {
	normalized := (Config{DSN: " postgres://user:pass@localhost:5432/app "}).normalized()
	if normalized.DSN != "postgres://user:pass@localhost:5432/app" {
		t.Fatalf("expected normalized dsn to be trimmed, got %q", normalized.DSN)
	}
	if normalized.MaxOpenConns != defaultMaxOpenConns {
		t.Fatalf("expected default max open conns %d, got %d", defaultMaxOpenConns, normalized.MaxOpenConns)
	}
	if normalized.MaxIdleConns != defaultMaxIdleConns {
		t.Fatalf("expected default max idle conns %d, got %d", defaultMaxIdleConns, normalized.MaxIdleConns)
	}
	if normalized.ConnMaxLifetime != defaultConnMaxLifetime {
		t.Fatalf("expected default conn max lifetime %s, got %s", defaultConnMaxLifetime, normalized.ConnMaxLifetime)
	}
}
