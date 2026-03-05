package agent

import "testing"

func TestStripCopilotUsageFooter_RemovesTelemetryBlock(t *testing.T) {
	input := `Ingest the taskboard by creating a prioritized backlog.

Total usage est:        1 Premium request
API time spent:         2s
Total session time:     4s
Total code changes:     +0 -0
Breakdown by AI model:
 gpt-5.3-codex           14.8k in, 149 out`

	got := stripCopilotUsageFooter(input)
	want := "Ingest the taskboard by creating a prioritized backlog."
	if got != want {
		t.Fatalf("stripCopilotUsageFooter() = %q, want %q", got, want)
	}
}

func TestStripCopilotUsageFooter_LeavesPlainPromptUntouched(t *testing.T) {
	input := "Create a concise taskboard ingestion prompt with explicit scope and priorities."
	got := stripCopilotUsageFooter(input)
	if got != input {
		t.Fatalf("stripCopilotUsageFooter() changed prompt: got %q", got)
	}
}

func TestStripCopilotUsageFooter_HandlesLeadingFooterMarker(t *testing.T) {
	input := "Total usage est: 1 Premium request"
	got := stripCopilotUsageFooter(input)
	if got != "" {
		t.Fatalf("stripCopilotUsageFooter() = %q, want empty string", got)
	}
}
