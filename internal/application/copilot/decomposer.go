package copilot

import "context"

type DecomposeRequest struct {
	RunID            string
	Prompt           string
	Model            string
	WorkingDirectory string
	SkillDirectories []string
	GitHubToken      string
	CLIPath          string
	CLIURL           string
}

type DecomposeResult struct {
	RunID      string
	SessionID  string
	Response   string
	Model      string
	PromptHash string
}

type Decomposer interface {
	Decompose(ctx context.Context, request DecomposeRequest) (DecomposeResult, error)
}
