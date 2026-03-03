package ingestion

import (
	"agentic-orchestrator/internal/domain/failures"
	domaintracker "agentic-orchestrator/internal/domain/tracker"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Document struct {
	ProjectID   string
	DocumentID  string
	FileName    string
	ContentType string
	ObjectPath  string
	CDNURL      string
	Status      string
}

type BoardStore interface {
	UpsertBoard(ctx context.Context, board domaintracker.Board) error
}

type ArtifactFetcher interface {
	FetchToPath(ctx context.Context, objectPath string, destinationPath string) error
}

type ArtifactFetcherFunc func(ctx context.Context, objectPath string, destinationPath string) error

func (function ArtifactFetcherFunc) FetchToPath(ctx context.Context, objectPath string, destinationPath string) error {
	return function(ctx, objectPath, destinationPath)
}

type Request struct {
	RunID                     string
	ProjectID                 string
	BoardID                   string
	SelectedDocumentLocations []string
	SystemPrompt              string
	UserPrompt                string
}

func (request Request) Validate() error {
	if strings.TrimSpace(request.RunID) == "" {
		return failures.WrapTerminal(errors.New("run_id is required"))
	}
	if strings.TrimSpace(request.ProjectID) == "" {
		return failures.WrapTerminal(errors.New("project_id is required"))
	}
	if len(request.SelectedDocumentLocations) == 0 {
		return failures.WrapTerminal(errors.New("selected_document_locations are required"))
	}
	if strings.TrimSpace(request.SystemPrompt) == "" {
		return failures.WrapTerminal(errors.New("system_prompt is required"))
	}
	return nil
}

type Service struct {
	boardStore      BoardStore
	artifactFetcher ArtifactFetcher
}

func NewService(boardStore BoardStore, artifactFetcher ArtifactFetcher) (*Service, error) {
	if boardStore == nil {
		return nil, failures.WrapTerminal(errors.New("ingestion board store is required"))
	}
	if artifactFetcher == nil {
		return nil, failures.WrapTerminal(errors.New("ingestion artifact fetcher is required"))
	}
	return &Service{boardStore: boardStore, artifactFetcher: artifactFetcher}, nil
}

func (service *Service) Execute(ctx context.Context, request Request) (domaintracker.Board, error) {
	if service == nil {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("ingestion service is not initialized"))
	}
	if err := request.Validate(); err != nil {
		return domaintracker.Board{}, err
	}
	boardID := strings.TrimSpace(request.BoardID)
	if boardID == "" {
		boardID = defaultBoardID(request.ProjectID)
	}
	now := time.Now().UTC()
	tempDir, err := os.MkdirTemp("", "ingestion-artifacts-")
	if err != nil {
		return domaintracker.Board{}, failures.WrapTransient(fmt.Errorf("create ingestion temp directory: %w", err))
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	hasDocument := false
	for index, rawDocumentLocation := range request.SelectedDocumentLocations {
		documentLocation := strings.TrimSpace(rawDocumentLocation)
		if documentLocation == "" {
			continue
		}
		hasDocument = true
		fileName := filepath.Base(documentLocation)
		localFilePath := filepath.Join(tempDir, fmt.Sprintf("%d_%s", index+1, sanitizeLocalFileName(fileName)))
		if fetchErr := service.artifactFetcher.FetchToPath(ctx, documentLocation, localFilePath); fetchErr != nil {
			return domaintracker.Board{}, ensureClassified(fetchErr)
		}
	}
	if !hasDocument {
		return domaintracker.Board{}, failures.WrapTerminal(errors.New("selected_document_locations must contain at least one non-empty value"))
	}
	_ = boardID
	_ = now
	panic("TODO: IMPLEMENT AGENTIC AGENT")
}

func ensureClassified(err error) error {
	if err == nil {
		return nil
	}
	if failures.ClassOf(err) != failures.ClassUnknown {
		return err
	}
	return failures.WrapTransient(err)
}

var nonAlphaNumericSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func sanitizeLocalFileName(fileName string) string {
	trimmed := strings.TrimSpace(fileName)
	if trimmed == "" {
		return "document.txt"
	}
	trimmed = strings.ReplaceAll(trimmed, string(filepath.Separator), "_")
	trimmed = strings.ReplaceAll(trimmed, "/", "_")
	return trimmed
}

func defaultBoardID(projectID string) string {
	cleanProjectID := strings.ToLower(strings.TrimSpace(projectID))
	cleanProjectID = nonAlphaNumericSanitizer.ReplaceAllString(cleanProjectID, "_")
	cleanProjectID = strings.Trim(cleanProjectID, "_")
	if cleanProjectID == "" {
		cleanProjectID = "project"
	}
	return fmt.Sprintf("%s_ingestion", cleanProjectID)
}
