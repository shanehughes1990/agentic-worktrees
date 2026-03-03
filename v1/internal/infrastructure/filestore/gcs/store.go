package gcs

import (
	applicationcontrolplane "agentic-orchestrator/internal/application/controlplane"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"

	"google.golang.org/api/option"
)

type Config struct {
	Bucket             string
	ServiceAccountJSON string
	RootPrefix         string
}

type Store struct {
	bucket        string
	rootPrefix    string
	accessID      string
	privateKey    []byte
	operationLock sync.Mutex
	storageClient *storage.Client
}

type serviceAccountCredential struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
}

func NewStore(ctx context.Context, config Config) (*Store, error) {
	bucket := strings.TrimSpace(config.Bucket)
	if bucket == "" {
		return nil, fmt.Errorf("gcs bucket is required")
	}
	credentialJSON := strings.TrimSpace(config.ServiceAccountJSON)
	if credentialJSON == "" {
		return nil, fmt.Errorf("gcs service account json is required")
	}
	var credential serviceAccountCredential
	if err := json.Unmarshal([]byte(credentialJSON), &credential); err != nil {
		return nil, fmt.Errorf("decode gcs service account json: %w", err)
	}
	if strings.TrimSpace(credential.ClientEmail) == "" {
		return nil, fmt.Errorf("gcs service account client_email is required")
	}
	if strings.TrimSpace(credential.PrivateKey) == "" {
		return nil, fmt.Errorf("gcs service account private_key is required")
	}
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(credentialJSON)))
	if err != nil {
		return nil, fmt.Errorf("create gcs storage client: %w", err)
	}
	rootPrefix := strings.Trim(path.Clean(strings.TrimSpace(config.RootPrefix)), "/")
	if rootPrefix == "" || rootPrefix == "." {
		rootPrefix = "projects"
	}
	return &Store{
		bucket:      bucket,
		rootPrefix:  rootPrefix,
		accessID:    strings.TrimSpace(credential.ClientEmail),
		privateKey:  []byte(credential.PrivateKey),
		storageClient: client,
	}, nil
}

func (store *Store) CreateSignedUploadURL(ctx context.Context, objectPath string, contentType string, expiresAt time.Time) (string, error) {
	_ = ctx
	if store == nil {
		return "", fmt.Errorf("gcs store is not initialized")
	}
	store.operationLock.Lock()
	defer store.operationLock.Unlock()
	cleanObjectPath, err := store.lockedObjectPath(objectPath)
	if err != nil {
		return "", err
	}
	trimmedContentType := strings.TrimSpace(contentType)
	if trimmedContentType == "" {
		trimmedContentType = "application/octet-stream"
	}
	if expiresAt.IsZero() {
		expiresAt = time.Now().UTC().Add(15 * time.Minute)
	}
	signedURL, err := storage.SignedURL(store.bucket, cleanObjectPath, &storage.SignedURLOptions{
		GoogleAccessID: store.accessID,
		PrivateKey:     store.privateKey,
		Method:         "PUT",
		ContentType:    trimmedContentType,
		Expires:        expiresAt.UTC(),
	})
	if err != nil {
		return "", fmt.Errorf("create gcs signed upload url: %w", err)
	}
	return signedURL, nil
}

func (store *Store) DeleteObject(ctx context.Context, objectPath string) error {
	if store == nil || store.storageClient == nil {
		return fmt.Errorf("gcs store is not initialized")
	}
	store.operationLock.Lock()
	defer store.operationLock.Unlock()
	cleanObjectPath, err := store.lockedObjectPath(objectPath)
	if err != nil {
		return err
	}
	if err := store.storageClient.Bucket(store.bucket).Object(cleanObjectPath).Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return nil
		}
		return fmt.Errorf("delete gcs object: %w", err)
	}
	return nil
}

func (store *Store) DownloadObjectToFile(ctx context.Context, objectPath string, destinationPath string) error {
	if store == nil || store.storageClient == nil {
		return fmt.Errorf("gcs store is not initialized")
	}
	store.operationLock.Lock()
	defer store.operationLock.Unlock()
	cleanObjectPath, err := store.lockedObjectPath(objectPath)
	if err != nil {
		return err
	}
	trimmedDestinationPath := strings.TrimSpace(destinationPath)
	if trimmedDestinationPath == "" {
		return fmt.Errorf("destination_path is required")
	}
	if err := os.MkdirAll(filepath.Dir(trimmedDestinationPath), 0o755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}
	reader, err := store.storageClient.Bucket(store.bucket).Object(cleanObjectPath).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("open gcs object reader: %w", err)
	}
	defer reader.Close()
	outputFile, err := os.Create(trimmedDestinationPath)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer outputFile.Close()
	if _, err := io.Copy(outputFile, reader); err != nil {
		return fmt.Errorf("copy gcs object to destination file: %w", err)
	}
	return nil
}

func (store *Store) lockedObjectPath(objectPath string) (string, error) {
	cleanObjectPath := strings.Trim(path.Clean(strings.TrimSpace(objectPath)), "/")
	if cleanObjectPath == "" {
		return "", fmt.Errorf("object_path is required")
	}
	if !strings.HasPrefix(cleanObjectPath, store.rootPrefix+"/") && cleanObjectPath != store.rootPrefix {
		return "", fmt.Errorf("object_path %q is outside configured filestore root %q", cleanObjectPath, store.rootPrefix)
	}
	return cleanObjectPath, nil
}

var _ applicationcontrolplane.ProjectFileStore = (*Store)(nil)
