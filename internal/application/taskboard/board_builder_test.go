package taskboard

import "testing"

func TestBuildBoardFromResponse(t *testing.T) {
	response := `{"status":"in-progress","epics":[{"id":"e1","title":"Epic","status":"completed","tasks":[{"id":"t1","title":"Task","status":"not-started"}]}],"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}`

	board, err := BuildBoardFromResponse("run-1", response)
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if board.BoardID != "run-1" || board.RunID != "run-1" {
		t.Fatalf("expected run ids to be applied, got board_id=%s run_id=%s", board.BoardID, board.RunID)
	}
}

func TestBuildBoardFromResponseRejectsInvalidPayload(t *testing.T) {
	_, err := BuildBoardFromResponse("run-1", "not-json")
	if err == nil {
		t.Fatalf("expected parse error")
	}
}
