package taskboard

import (
	"fmt"
	"sort"
	"strings"

	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type Navigator struct{}

func NewNavigator() *Navigator {
	return &Navigator{}
}

func (navigator *Navigator) GetNextTask(board *Board) (*Task, error) {
	readyTasks, err := navigator.GetReadyTasks(board)
	if err != nil {
		return nil, err
	}
	if len(readyTasks) == 0 {
		return nil, nil
	}
	return readyTasks[0], nil
}

func (navigator *Navigator) GetReadyTasks(board *Board) ([]*Task, error) {
	epicOrder, err := navigator.buildEpicOrder(board)
	if err != nil {
		return nil, err
	}

	taskOrder, incomingTaskEdges, err := navigator.buildTaskOrder(board)
	if err != nil {
		return nil, err
	}

	epicRank := map[string]int{}
	for index, epicID := range epicOrder {
		epicRank[epicID] = index
	}
	taskRank := map[string]int{}
	for index, taskID := range taskOrder {
		taskRank[taskID] = index
	}

	taskToEpic := map[string]string{}
	readyTasks := make([]*Task, 0)
	for epicIndex := range board.Epics {
		epic := &board.Epics[epicIndex]
		for taskIndex := range epic.Tasks {
			task := &epic.Tasks[taskIndex]
			taskToEpic[task.ID] = epic.ID
			if task.Status == StatusCompleted || task.Status == StatusInProgress {
				continue
			}

			allPrerequisitesCompleted := true
			for _, prerequisiteTaskID := range incomingTaskEdges[task.ID] {
				if !board.IsCompleted(prerequisiteTaskID) {
					allPrerequisitesCompleted = false
					break
				}
			}
			if allPrerequisitesCompleted {
				readyTasks = append(readyTasks, task)
			}
		}
	}

	sort.SliceStable(readyTasks, func(left, right int) bool {
		leftTask := readyTasks[left]
		rightTask := readyTasks[right]
		leftEpicRank := epicRank[taskToEpic[leftTask.ID]]
		rightEpicRank := epicRank[taskToEpic[rightTask.ID]]
		if leftEpicRank != rightEpicRank {
			return leftEpicRank < rightEpicRank
		}
		leftTaskRank := taskRank[leftTask.ID]
		rightTaskRank := taskRank[rightTask.ID]
		if leftTaskRank != rightTaskRank {
			return leftTaskRank < rightTaskRank
		}
		return leftTask.ID < rightTask.ID
	})

	return readyTasks, nil
}

func (navigator *Navigator) buildEpicOrder(board *Board) ([]string, error) {
	epicGraph := simple.NewDirectedGraph()
	epicToNumericID := map[string]int64{}
	numericToEpicID := map[int64]string{}
	nextID := int64(1)

	for _, epic := range board.Epics {
		epicID := strings.TrimSpace(epic.ID)
		epicToNumericID[epicID] = nextID
		numericToEpicID[nextID] = epicID
		epicGraph.AddNode(simple.Node(nextID))
		nextID++
	}

	for _, epic := range board.Epics {
		for _, dependencyEpicID := range epic.DependsOn {
			cleanDependencyEpicID := strings.TrimSpace(dependencyEpicID)
			if cleanDependencyEpicID == "" {
				continue
			}
			fromID := epicToNumericID[cleanDependencyEpicID]
			toID := epicToNumericID[epic.ID]
			epicGraph.SetEdge(epicGraph.NewEdge(simple.Node(fromID), simple.Node(toID)))
		}
	}

	nodes, err := topo.Sort(epicGraph)
	if err != nil {
		return nil, fmt.Errorf("invalid epic dependency graph: %w", err)
	}

	orderedEpicIDs := make([]string, 0, len(nodes))
	for _, node := range nodes {
		orderedEpicIDs = append(orderedEpicIDs, numericToEpicID[node.ID()])
	}
	return orderedEpicIDs, nil
}

func (navigator *Navigator) buildTaskOrder(board *Board) ([]string, map[string][]string, error) {
	if err := board.ValidateBasics(); err != nil {
		return nil, nil, err
	}

	taskGraph := simple.NewDirectedGraph()
	taskToNumericID := map[string]int64{}
	numericToTaskID := map[int64]string{}
	nextID := int64(1)

	for _, epic := range board.Epics {
		for _, task := range epic.Tasks {
			taskToNumericID[task.ID] = nextID
			numericToTaskID[nextID] = task.ID
			taskGraph.AddNode(simple.Node(nextID))
			nextID++
		}
	}

	incomingEdges := map[string][]string{}
	for _, epic := range board.Epics {
		for _, task := range epic.Tasks {
			for _, dependencyTaskID := range task.DependsOn {
				cleanDependencyTaskID := strings.TrimSpace(dependencyTaskID)
				if cleanDependencyTaskID == "" {
					continue
				}
				fromID := taskToNumericID[cleanDependencyTaskID]
				toID := taskToNumericID[task.ID]
				taskGraph.SetEdge(taskGraph.NewEdge(simple.Node(fromID), simple.Node(toID)))
				incomingEdges[task.ID] = append(incomingEdges[task.ID], cleanDependencyTaskID)
			}
		}
	}

	nodes, err := topo.Sort(taskGraph)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid task dependency graph: %w", err)
	}

	orderedTaskIDs := make([]string, 0, len(nodes))
	for _, node := range nodes {
		orderedTaskIDs = append(orderedTaskIDs, numericToTaskID[node.ID()])
	}

	return orderedTaskIDs, incomingEdges, nil
}
