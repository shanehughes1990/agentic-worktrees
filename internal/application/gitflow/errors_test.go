package gitflow

import (
	"fmt"
	"testing"
)

func TestIsTransientInfrastructureFailureStartupProbeKilled(t *testing.T) {
	err := fmt.Errorf("implement task with agent: copilot preflight failed: copilot cli startup probe failed: signal: killed")
	if !IsTransientInfrastructureFailure(err) {
		t.Fatalf("expected startup probe killed failure to be transient")
	}
}

func TestIsTransientInfrastructureFailureIgnoresTerminalAuthFailure(t *testing.T) {
	err := fmt.Errorf("start copilot client: authentication failed")
	if IsTransientInfrastructureFailure(err) {
		t.Fatalf("expected auth failure not to be treated as transient infra failure")
	}
}
