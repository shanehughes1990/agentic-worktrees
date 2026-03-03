package local

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Config struct {
	RootDirectory string
}

type Store struct {
	rootDirectory string
	operationLock sync.Mutex
}

func NewStore(config Config) (*Store, error) {
	root := filepath.Clean(strings.TrimSpace(config.RootDirectory))
	if root == "" || root == "." {
		return nil, fmt.Errorf("local filestore root_directory is required")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create local filestore root: %w", err)
	}
	return &Store{rootDirectory: root}, nil
}

func (store *Store) CreateSignedUploadURL(ctx context.Context, objectPath string, contentType string, expiresAt time.Time) (string, error) {
	_, _, _ = ctx, contentType, expiresAt
	if store == nil {
		return "", fmt.Errorf("local filestore is not initialized")
	}
	store.operationLock.Lock()
	defer store.operationLock.Unlock()
	resolved, err := store.lockedResolveObjectPath(objectPath)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return "", fmt.Errorf("ensure local upload directory: %w", err)
	}
	return "file://" + resolved, nil
}

func (store *Store) DeleteObject(ctx context.Context, objectPath string) error {
	_ = ctx
	if store == nil {
		return fmt.Errorf("local filestore is not initialized")
	}
	store.operationLock.Lock()
	defer store.operationLock.Unlock()
	resolved, err := store.lockedResolveObjectPath(objectPath)
	if err != nil {
		return err
	}
	err = os.Remove(resolved)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete local object: %w", err)
	}
	return nil
}

func (store *Store) ArtifactsRoot(projectID string) (string, error) {
	if store == nil {
		return "", fmt.Errorf("local filestore is not initialized")
	}
	store.operationLock.Lock()
	defer store.operationLock.Unlock()
	return store.lockedResolveArtifactDir(projectID, "artifacts")
}

func (store *Store) SourceReposRoot(projectID string) (string, error) {
	if store == nil {
		return "", fmt.Errorf("local filestore is not initialized")
	}
	store.operationLock.Lock()
	defer store.operationLock.Unlock()
	return store.lockedResolveArtifactDir(projectID, "repositories", "source")
}

func (store *Store) WorktreesRoot(projectID string) (string, error) {
	if store == nil {
		return "", fmt.Errorf("local filestore is not initialized")
	}
	store.operationLock.Lock()
	defer store.operationLock.Unlock()
	return store.lockedResolveArtifactDir(projectID, "worktrees")
}

func (store *Store) EnsureArtifactSubfolders(projectID string) error {
	if store == nil {
		return fmt.Errorf("local filestore is not initialized")
	}
	store.operationLock.Lock()
	defer store.operationLock.Unlock()
	directories := [][]string{
		{"artifacts"},
		{"repositories", "source"},
		{"worktrees"},
		{"tracker"},
		{"logs"},
	}
	for _, segments := range directories {
		directory, err := store.lockedResolveArtifactDir(projectID, segments...)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return fmt.Errorf("create artifact subfolder %q: %w", filepath.Join(segments...), err)
		}
	}
	return nil
}

func (store *Store) lockedResolveObjectPath(objectPath string) (string, error) {
	cleanObjectPath := filepath.Clean(strings.TrimSpace(objectPath))
	if cleanObjectPath == "" || cleanObjectPath == "." {
		return "", fmt.Errorf("object_path is required")
	}
	resolved := filepath.Clean(filepath.Join(store.rootDirectory, cleanObjectPath))
	if !isWithinRoot(resolved, store.rootDirectory) {
		return "", fmt.Errorf("object_path %q is outside local filestore root", objectPath)
	}
	return resolved, nil
}

func (store *Store) lockedResolveArtifactDir(projectID string, segments ...string) (string, error) {
	cleanProjectID := sanitizeProjectID(projectID)
	parts := []string{"projects", cleanProjectID}
	parts = append(parts, segments...)
	directory := filepath.Clean(filepath.Join(append([]string{store.rootDirectory}, parts...)...))
	if !isWithinRoot(directory, store.rootDirectory) {
		return "", fmt.Errorf("artifact directory is outside local filestore root")
	}
	return directory, nil
}

func sanitizeProjectID(projectID string) string {
	clean := strings.TrimSpace(projectID)
	if clean == "" {
		return "unscoped"
	}
	clean = strings.ReplaceAll(clean, "..", "_")
	clean = strings.ReplaceAll(clean, string(filepath.Separator), "_")
	clean = strings.ReplaceAll(clean, "/", "_")
	return clean
}

func isWithinRoot(candidate string, root string) bool {
	cleanRoot := filepath.Clean(root)
	cleanCandidate := filepath.Clean(candidate)
	if cleanCandidate == cleanRoot {
		return true
	}
	return strings.HasPrefix(cleanCandidate, cleanRoot+string(filepath.Separator))
}

var _ applicationcontrolplane.ProjectFileStore = (*Store)(nil)
