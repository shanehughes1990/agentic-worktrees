package copilot

import "strings"

const defaultModel = "gpt-5.3-codex"

type ClientConfig struct {
	GitHubToken       string
	CLIPath           string
	CLIURL            string
	AuthStatusCommand string
	AuthLoginCommand  string
	LogLevel          string
	DefaultModel      string
	SkillDirectories  []string
}

func (cfg ClientConfig) Normalized() ClientConfig {
	normalized := cfg
	normalized.GitHubToken = strings.TrimSpace(normalized.GitHubToken)
	normalized.CLIPath = strings.TrimSpace(normalized.CLIPath)
	normalized.CLIURL = strings.TrimSpace(normalized.CLIURL)
	normalized.AuthStatusCommand = strings.TrimSpace(normalized.AuthStatusCommand)
	if normalized.AuthStatusCommand == "" {
		normalized.AuthStatusCommand = "copilot auth status"
	}
	normalized.AuthLoginCommand = strings.TrimSpace(normalized.AuthLoginCommand)
	if normalized.AuthLoginCommand == "" {
		normalized.AuthLoginCommand = "copilot auth login"
	}
	normalized.LogLevel = strings.TrimSpace(normalized.LogLevel)
	if normalized.LogLevel == "" {
		normalized.LogLevel = "error"
	}
	normalized.DefaultModel = strings.TrimSpace(normalized.DefaultModel)
	if normalized.DefaultModel == "" {
		normalized.DefaultModel = defaultModel
	}
	return normalized
}
