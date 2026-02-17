package app

import "testing"

func TestCheckQueueName(t *testing.T) {
	if checkQueueName("default").OK != true {
		t.Fatalf("expected queue check to pass")
	}
	if checkQueueName("   ").OK != false {
		t.Fatalf("expected empty queue to fail")
	}
}

func TestCheckPathsAndADKURL(t *testing.T) {
	if checkBoardPath("state/board.json").OK != true {
		t.Fatalf("expected board path check to pass")
	}
	if checkBoardPath("").OK != false {
		t.Fatalf("expected empty board path to fail")
	}
	if checkCheckpointPath("state/checkpoints.json").OK != true {
		t.Fatalf("expected checkpoint path check to pass")
	}
	if checkCheckpointPath(" ").OK != false {
		t.Fatalf("expected empty checkpoint path to fail")
	}
	if checkADKURL("https://example.invalid").OK != true {
		t.Fatalf("expected adk url check to pass")
	}
	if checkADKURL(" ").OK != false {
		t.Fatalf("expected empty adk url to fail")
	}
}

func TestPreflightReportError(t *testing.T) {
	report := PreflightReport{
		Healthy: false,
		Checks: []CheckResult{
			{Name: "redis", OK: true, Detail: "reachable"},
			{Name: "copilot_adk_board_url", OK: false, Detail: "missing"},
		},
	}
	if err := report.Error(); err == nil {
		t.Fatalf("expected report error")
	}

	ok := PreflightReport{Healthy: true}
	if err := ok.Error(); err != nil {
		t.Fatalf("expected nil error for healthy report, got %v", err)
	}
}
