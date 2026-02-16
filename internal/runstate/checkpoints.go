package runstate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Checkpoint struct {
	RunID        string    `json:"run_id"`
	TaskID       string    `json:"task_id"`
	Phase        string    `json:"phase"`
	Status       string    `json:"status"`
	OriginBranch string    `json:"origin_branch,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type fileModel struct {
	SchemaVersion int          `json:"schema_version"`
	Checkpoints   []Checkpoint `json:"checkpoints"`
}

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Upsert(checkpoint Checkpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	model, err := s.loadLocked()
	if err != nil {
		return err
	}

	checkpoint.UpdatedAt = time.Now().UTC()
	updated := false
	for i := range model.Checkpoints {
		if model.Checkpoints[i].RunID == checkpoint.RunID && model.Checkpoints[i].TaskID == checkpoint.TaskID {
			model.Checkpoints[i] = checkpoint
			updated = true
			break
		}
	}
	if !updated {
		model.Checkpoints = append(model.Checkpoints, checkpoint)
	}

	return s.saveLocked(model)
}

func (s *Store) All() ([]Checkpoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	model, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	result := make([]Checkpoint, len(model.Checkpoints))
	copy(result, model.Checkpoints)
	return result, nil
}

func (s *Store) Summary() (map[string]int, error) {
	checkpoints, err := s.All()
	if err != nil {
		return nil, err
	}
	summary := map[string]int{}
	for _, checkpoint := range checkpoints {
		summary[checkpoint.Status]++
	}
	return summary, nil
}

func (s *Store) loadLocked() (fileModel, error) {
	payload, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return fileModel{SchemaVersion: 1, Checkpoints: []Checkpoint{}}, nil
		}
		return fileModel{}, fmt.Errorf("read checkpoint file: %w", err)
	}

	var model fileModel
	if err := json.Unmarshal(payload, &model); err != nil {
		return fileModel{}, fmt.Errorf("unmarshal checkpoint file: %w", err)
	}
	if model.SchemaVersion <= 0 {
		model.SchemaVersion = 1
	}
	if model.Checkpoints == nil {
		model.Checkpoints = []Checkpoint{}
	}
	return model, nil
}

func (s *Store) saveLocked(model fileModel) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create checkpoint dir: %w", err)
	}
	payload, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal checkpoint model: %w", err)
	}
	payload = append(payload, '\n')
	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0o644); err != nil {
		return fmt.Errorf("write tmp checkpoint file: %w", err)
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("rename checkpoint file: %w", err)
	}
	return nil
}
