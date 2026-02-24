package taskboard

import (
	"fmt"
	"sort"

	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type Navigator struct{}

func NewNavigator() *Navigator {
	return &Navigator{}
}

func (navigator *Navigator) GetNextTask(board *Board) (*MicroTask, error) {
	readyTasks, err := navigator.GetReadyTasks(board)
	if err != nil {
		return nil, err
	}
	if len(readyTasks) == 0 {
		return nil, nil
	}
	return readyTasks[0], nil
}

func (navigator *Navigator) GetReadyTasks(board *Board) ([]*MicroTask, error) {
	orderedNodeIDs, incomingEdges, err := navigator.buildDAG(board)
	if err != nil {
		return nil, err
	}

	microTaskByID := make(map[string]*MicroTask, len(board.MicroTasks))
	for index := range board.MicroTasks {
		microTask := &board.MicroTasks[index]
		microTaskByID[microTask.ID] = microTask
	}

	readyTasks := make([]*MicroTask, 0)
	orderIndex := make(map[string]int, len(orderedNodeIDs))
	for index, nodeID := range orderedNodeIDs {
		orderIndex[nodeID] = index
	}

	for _, microTask := range board.MicroTasks {
		if microTask.Status == StatusCompleted || microTask.Status == StatusInProgress {
			continue
		}

		prerequisites := incomingEdges[microTask.ID]
		allPrerequisitesCompleted := true
		for _, prerequisiteID := range prerequisites {
			if !board.IsCompleted(prerequisiteID) {
				allPrerequisitesCompleted = false
				break
			}
		}
		if allPrerequisitesCompleted {
			readyTasks = append(readyTasks, microTaskByID[microTask.ID])
		}
	}

	sort.SliceStable(readyTasks, func(left, right int) bool {
		leftIndex := orderIndex[readyTasks[left].ID]
		rightIndex := orderIndex[readyTasks[right].ID]
		if leftIndex == rightIndex {
			return readyTasks[left].ID < readyTasks[right].ID
		}
		return leftIndex < rightIndex
	})

	return readyTasks, nil
}

func (navigator *Navigator) buildDAG(board *Board) ([]string, map[string][]string, error) {
	if err := board.ValidateBasics(); err != nil {
		return nil, nil, err
	}

	taskGraph := simple.NewDirectedGraph()
	nodeToNumericID := make(map[string]int64)
	numericToNodeID := make(map[int64]string)
	nextID := int64(1)

	addNode := func(workItemID string) {
		nodeToNumericID[workItemID] = nextID
		numericToNodeID[nextID] = workItemID
		taskGraph.AddNode(simple.Node(nextID))
		nextID++
	}

	for _, epic := range board.Epics {
		addNode(epic.ID)
	}
	for _, task := range board.Tasks {
		addNode(task.ID)
	}
	for _, microTask := range board.MicroTasks {
		addNode(microTask.ID)
	}

	incomingEdges := make(map[string][]string)
	for _, dependency := range board.Dependencies {
		fromNumericID := nodeToNumericID[dependency.FromID]
		toNumericID := nodeToNumericID[dependency.ToID]
		taskGraph.SetEdge(taskGraph.NewEdge(simple.Node(fromNumericID), simple.Node(toNumericID)))
		incomingEdges[dependency.ToID] = append(incomingEdges[dependency.ToID], dependency.FromID)
	}

	nodes, err := topo.Sort(taskGraph)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid dependency graph: %w", err)
	}

	orderedNodeIDs := make([]string, 0, len(nodes))
	for _, node := range nodes {
		orderedNodeIDs = append(orderedNodeIDs, numericToNodeID[node.ID()])
	}

	return orderedNodeIDs, incomingEdges, nil
}
