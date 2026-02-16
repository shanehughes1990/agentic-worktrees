package board

import "testing"

func TestBuildBoardFromContent(t *testing.T) {
	content := "- Implement parser [depends:task-010]\n- Add tests"
	board := BuildBoardFromContent("scope.md", content)
	if board.SchemaVersion != 1 {
		t.Fatalf("expected schema version 1, got %d", board.SchemaVersion)
	}
	if len(board.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(board.Tasks))
	}
	if board.Tasks[0].ID != "task-001" {
		t.Fatalf("unexpected first task id: %s", board.Tasks[0].ID)
	}
	if len(board.Tasks[0].Dependencies) != 1 || board.Tasks[0].Dependencies[0] != "task-010" {
		t.Fatalf("unexpected dependencies: %+v", board.Tasks[0].Dependencies)
	}
}
