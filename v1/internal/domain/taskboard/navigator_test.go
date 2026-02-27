package taskboard

import "testing"

func TestNavigatorGetNextTask(t *testing.T) {
	board := validBoard()
	board.Epics[0].Tasks[0].Status = StatusNotStarted
	board.Epics[0].Tasks = append(board.Epics[0].Tasks,
		Task{WorkItem: WorkItem{ID: "t2", BoardID: "board-1", Title: "T2", Status: StatusNotStarted}, DependsOn: []string{"t1"}},
	)

	navigator := NewNavigator()
	nextTask, err := navigator.GetNextTask(board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nextTask == nil || nextTask.ID != "t1" {
		t.Fatalf("expected t1 as next task, got %#v", nextTask)
	}

	if err := board.SetTaskStatus("t1", StatusCompleted); err != nil {
		t.Fatalf("unexpected status update error: %v", err)
	}
	nextTask, err = navigator.GetNextTask(board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nextTask == nil || nextTask.ID != "t2" {
		t.Fatalf("expected t2 as next task, got %#v", nextTask)
	}
}

func TestNavigatorConcurrentReadyTasks(t *testing.T) {
	board := validBoard()
	board.Epics[0].Tasks = []Task{
		{WorkItem: WorkItem{ID: "t1", BoardID: "board-1", Title: "T1", Status: StatusNotStarted}},
		{WorkItem: WorkItem{ID: "t2", BoardID: "board-1", Title: "T2", Status: StatusNotStarted}},
	}

	navigator := NewNavigator()
	readyTasks, err := navigator.GetReadyTasks(board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(readyTasks) != 2 {
		t.Fatalf("expected two concurrent ready tasks, got %d", len(readyTasks))
	}
}

func TestNavigatorRejectsCycle(t *testing.T) {
	board := validBoard()
	board.Epics[0].Tasks = []Task{
		{WorkItem: WorkItem{ID: "t1", BoardID: "board-1", Title: "T1", Status: StatusNotStarted}, DependsOn: []string{"t2"}},
		{WorkItem: WorkItem{ID: "t2", BoardID: "board-1", Title: "T2", Status: StatusNotStarted}, DependsOn: []string{"t1"}},
	}

	navigator := NewNavigator()
	if _, err := navigator.GetReadyTasks(board); err == nil {
		t.Fatalf("expected cycle error")
	}
}
