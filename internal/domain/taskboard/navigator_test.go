package taskboard

import "testing"

func TestNavigatorGetNextTask(t *testing.T) {
	board := validBoard()
	board.Tasks[0].Status = StatusCompleted
	board.Epics[0].Status = StatusCompleted
	board.MicroTasks = []MicroTask{
		{WorkItem: WorkItem{ID: "m1", BoardID: "board-1", Title: "M1", Status: StatusNotStarted}, ItemType: "micro_task", TaskID: "t1"},
		{WorkItem: WorkItem{ID: "m2", BoardID: "board-1", Title: "M2", Status: StatusNotStarted}, ItemType: "micro_task", TaskID: "t1"},
	}
	board.Dependencies = []Dependency{
		{EdgeID: "d1", BoardID: "board-1", FromID: "t1", ToID: "m1"},
		{EdgeID: "d2", BoardID: "board-1", FromID: "m1", ToID: "m2"},
	}

	navigator := NewNavigator()
	nextTask, err := navigator.GetNextTask(board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nextTask == nil || nextTask.ID != "m1" {
		t.Fatalf("expected m1 as next task, got %#v", nextTask)
	}

	if err := board.SetMicroTaskStatus("m1", StatusCompleted); err != nil {
		t.Fatalf("unexpected status update error: %v", err)
	}
	nextTask, err = navigator.GetNextTask(board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nextTask == nil || nextTask.ID != "m2" {
		t.Fatalf("expected m2 as next task, got %#v", nextTask)
	}
}

func TestNavigatorRejectsCycle(t *testing.T) {
	board := validBoard()
	board.MicroTasks = append(board.MicroTasks, MicroTask{WorkItem: WorkItem{ID: "m2", BoardID: "board-1", Title: "M2", Status: StatusNotStarted}, ItemType: "micro_task", TaskID: "t1"})
	board.Dependencies = []Dependency{
		{EdgeID: "d1", BoardID: "board-1", FromID: "m1", ToID: "m2"},
		{EdgeID: "d2", BoardID: "board-1", FromID: "m2", ToID: "m1"},
	}

	navigator := NewNavigator()
	if _, err := navigator.GetReadyTasks(board); err == nil {
		t.Fatalf("expected cycle error")
	}
}
