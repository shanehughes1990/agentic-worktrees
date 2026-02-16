package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AuditSink struct {
	path string
	file *os.File
	mu   sync.Mutex
}

type AuditEvent struct {
	At      time.Time       `json:"at"`
	Type    string          `json:"type"`
	TaskID  string          `json:"task_id,omitempty"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func NewAuditSink(path string) (*AuditSink, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir audit dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open audit file: %w", err)
	}
	return &AuditSink{path: path, file: f}, nil
}

func (a *AuditSink) Path() string {
	return a.path
}

func (a *AuditSink) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.file == nil {
		return nil
	}
	err := a.file.Close()
	a.file = nil
	return err
}

func (a *AuditSink) Write(ctx context.Context, event AuditEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file == nil {
		return fmt.Errorf("audit sink is closed")
	}

	if event.At.IsZero() {
		event.At = time.Now().UTC()
	}

	line, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	if _, err = a.file.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}

	return nil
}
