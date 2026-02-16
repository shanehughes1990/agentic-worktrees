package gitops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Manager struct {
	root string
}

func NewManager(root string) *Manager {
	return &Manager{root: root}
}

func (m *Manager) EnsureWorktree(worktreeName string) (string, error) {
	name := strings.TrimSpace(worktreeName)
	if name == "" {
		return "", fmt.Errorf("worktree name cannot be empty")
	}
	path := filepath.Join(m.root, name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", fmt.Errorf("create worktree path: %w", err)
	}
	return path, nil
}

func (m *Manager) MarkMerged(runID string, taskID string, originBranch string) error {
	if err := os.MkdirAll("logs", 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile("logs/merge.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("run=%s task=%s branch=%s status=merged\n", runID, taskID, originBranch))
	return err
}
