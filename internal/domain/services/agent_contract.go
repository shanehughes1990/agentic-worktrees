package services

import "context"

type AgentRunner interface {
	DoTaskFromTaskBoard(ctx context.Context, request DoTaskFromTaskBoardRequest) (DoTaskFromTaskBoardResult, error)
	CreateTaskBoardFromTextFiles(ctx context.Context, request CreateTaskBoardFromTextFilesRequest) (CreateTaskBoardFromTextFilesResult, error)
	ResolveGitConflicts(ctx context.Context, request ResolveGitConflictsRequest) (ResolveGitConflictsResult, error)
}

type AgentRequestMetadata struct {
	RunID          string
	JobID          string
	Model          string
	RepositoryPath string
}

type DoTaskFromTaskBoardRequest struct {
	Metadata AgentRequestMetadata
	TaskID   string
	Prompt   string
}

type DoTaskFromTaskBoardResult struct {
	Summary      string
	ChangedFiles []string
}

type CreateTaskBoardFromTextFilesRequest struct {
	Metadata  AgentRequestMetadata
	FilePaths []string
	Prompt    string
}

type CreateTaskBoardFromTextFilesResult struct {
	BoardJSON string
}

type ResolveGitConflictsRequest struct {
	Metadata      AgentRequestMetadata
	ConflictFiles []string
	Prompt        string
}

type ResolveGitConflictsResult struct {
	ResolvedFiles []string
	Summary       string
}
