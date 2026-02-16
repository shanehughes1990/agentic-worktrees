package config

import "testing"

type testConfig struct {
	Name  string `envconfig:"NAME" validate:"required"`
	Port  int    `envconfig:"PORT" default:"8080" validate:"gte=1,lte=65535"`
	Debug bool   `envconfig:"DEBUG" default:"false"`
}

func TestDecodeSuccess(t *testing.T) {
	t.Setenv("APP_NAME", "agentic")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("APP_DEBUG", "true")

	var cfg testConfig
	if err := Decode("APP", &cfg); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Name != "agentic" {
		t.Fatalf("unexpected name: %q", cfg.Name)
	}
	if cfg.Port != 9090 {
		t.Fatalf("unexpected port: %d", cfg.Port)
	}
	if !cfg.Debug {
		t.Fatalf("expected debug=true")
	}
}

func TestDecodeValidationError(t *testing.T) {
	t.Setenv("APP_NAME", "")

	var cfg testConfig
	if err := Decode("APP", &cfg); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestDecodeNilTarget(t *testing.T) {
	if err := Decode("APP", nil); err == nil {
		t.Fatalf("expected nil target error")
	}
}

func TestLoadGeneric(t *testing.T) {
	t.Setenv("APP_NAME", "agentic")

	cfg, err := Load[testConfig]("APP")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Name != "agentic" {
		t.Fatalf("unexpected name: %q", cfg.Name)
	}
}
