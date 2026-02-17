package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	checkpointdomain "github.com/shanehughes1990/agentic-worktrees/internal/features/checkpoints/domain"
)

type JSONStore struct {
	path string
	mu   sync.Mutex
}

func NewJSONStore(path string) *JSONStore {
	return &JSONStore{path: path}
}

func (s *JSONStore) Append(record checkpointdomain.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record.Timestamp = record.Timestamp.UTC()
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now().UTC()
	}

	records, err := s.readAll()
	if err != nil {
		return err
	}
	records = append(records, record)
	return s.writeAll(records)
}

func (s *JSONStore) readAll() ([]checkpointdomain.Record, error) {
	payload, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []checkpointdomain.Record{}, nil
		}
		return nil, fmt.Errorf("read checkpoint file: %w", err)
	}

	if len(payload) == 0 {
		return []checkpointdomain.Record{}, nil
	}

	var records []checkpointdomain.Record
	if err := json.Unmarshal(payload, &records); err != nil {
		return nil, fmt.Errorf("unmarshal checkpoints: %w", err)
	}
	return records, nil
}

func (s *JSONStore) writeAll(records []checkpointdomain.Record) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create checkpoint dir: %w", err)
	}

	payload, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal checkpoints: %w", err)
	}
	payload = append(payload, '\n')

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, payload, 0o644); err != nil {
		return fmt.Errorf("write checkpoint tmp file: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("rename checkpoint tmp file: %w", err)
	}
	return nil
}
