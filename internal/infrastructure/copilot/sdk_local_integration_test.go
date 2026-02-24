package copilot

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	sdk "github.com/github/copilot-sdk/go"
	"golang.org/x/term"
)

func TestLocalCopilotSDKAuthStatus(t *testing.T) {
	if isCIEnvironment() {
		t.Skip("skipping local Copilot SDK integration test in CI")
	}
	if !isAllowedLocalExecution() {
		t.Skip("locked: run via VS Code play button or set RUN_LOCAL_COPILOT_SDK_TEST=1")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	compatCommand, err := runSDKCLICompatibilityCheck(ctx, "")
	if err != nil {
		if isSDKCLIIncompatibilityError(err) {
			t.Skipf("skipping local Copilot SDK integration test due to incompatible CLI/SDK launch contract. command=%q err=%v", compatCommand, err)
		}
		t.Fatalf("copilot sdk/cli compatibility check failed: command=%q err=%v", compatCommand, err)
	}

	client := sdk.NewClient(&sdk.ClientOptions{LogLevel: "error"})
	if err := client.Start(ctx); err != nil {
		t.Fatalf("start copilot sdk client: %v", err)
	}
	defer client.Stop()

	status, err := client.GetAuthStatus(ctx)
	if err != nil {
		t.Fatalf("get copilot sdk auth status: %v", err)
	}
	if status == nil {
		t.Fatalf("expected non-nil auth status")
	}
	if !status.IsAuthenticated {
		t.Fatalf("expected authenticated status, got isAuthenticated=false authType=%q login=%q host=%q statusMessage=%q", safeStringPtr(status.AuthType), safeStringPtr(status.Login), safeStringPtr(status.Host), safeStringPtr(status.StatusMessage))
	}
}

func isAllowedLocalExecution() bool {
	if strings.TrimSpace(os.Getenv("RUN_LOCAL_COPILOT_SDK_TEST")) == "1" {
		return true
	}
	return isVSCodePlayButtonContext()
}

func isVSCodePlayButtonContext() bool {
	if !isVSCodeEnvironment() {
		return false
	}
	return !term.IsTerminal(int(os.Stdout.Fd()))
}

func isVSCodeEnvironment() bool {
	markers := []string{"VSCODE_PID", "VSCODE_CWD", "VSCODE_IPC_HOOK_CLI"}
	for _, marker := range markers {
		if strings.TrimSpace(os.Getenv(marker)) != "" {
			return true
		}
	}
	return false
}

func isCIEnvironment() bool {
	keys := []string{
		"CI",
		"GITHUB_ACTIONS",
		"BUILDKITE",
		"JENKINS_URL",
		"TEAMCITY_VERSION",
		"TF_BUILD",
		"CIRCLECI",
		"TRAVIS",
	}
	for _, key := range keys {
		if strings.TrimSpace(os.Getenv(key)) != "" {
			return true
		}
	}
	return false
}
