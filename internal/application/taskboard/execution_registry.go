package taskboard

import (
	"strings"
	"sync"
)

type ExecutionRegistry struct {
	mu      sync.Mutex
	cancels map[string]func()
}

func NewExecutionRegistry() *ExecutionRegistry {
	return &ExecutionRegistry{cancels: map[string]func(){}}
}

func (registry *ExecutionRegistry) Register(boardID string, cancel func()) {
	cleanBoardID := strings.TrimSpace(boardID)
	if cleanBoardID == "" || cancel == nil {
		return
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.cancels[cleanBoardID] = cancel
}

func (registry *ExecutionRegistry) Unregister(boardID string) {
	cleanBoardID := strings.TrimSpace(boardID)
	if cleanBoardID == "" {
		return
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	delete(registry.cancels, cleanBoardID)
}

func (registry *ExecutionRegistry) Cancel(boardID string) bool {
	cleanBoardID := strings.TrimSpace(boardID)
	if cleanBoardID == "" {
		return false
	}
	registry.mu.Lock()
	cancel := registry.cancels[cleanBoardID]
	registry.mu.Unlock()
	if cancel == nil {
		return false
	}
	cancel()
	return true
}
