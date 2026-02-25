package dashboard

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
)

type IngestFunc func(ctx context.Context, directory string) (apptaskboard.IngestionResult, error)
type StartTaskTreeFunc func(ctx context.Context, boardID string, sourceBranch string, maxTasks int) (string, error)
type CancelTaskTreeFunc func(ctx context.Context, boardID string) (string, error)
type ListTaskboardsFunc func(ctx context.Context) ([]string, error)
type ListReadyTaskIDsFunc func(ctx context.Context, boardID string) ([]string, error)
type ListWorkflowsFunc func(ctx context.Context) ([]apptaskboard.IngestionWorkflow, error)
type WorkflowStatusFunc func(ctx context.Context, runID string) (*apptaskboard.IngestionWorkflow, error)
type CopilotAuthStatusFunc func(ctx context.Context) (string, error)
type CopilotAuthenticateFunc func(ctx context.Context) error

type Command struct {
	Title       string
	Description string
	Shortcut    rune
	Action      func()
}

type UI struct {
	application          *tview.Application
	pages                *tview.Pages
	status               *tview.TextView
	maxTaskLimit         int
	screenStack          []string
	currentScreen        string
	mainCommandList      *tview.List
	authCommandList      *tview.List
	ingestionCommandList *tview.List
	worktreeCommandList  *tview.List
	runIngestionInput    *tview.InputField
	runIngestionCommands *tview.List
	runGitBoardList      *tview.List
	runGitSourceInput    *tview.InputField
	runGitConcurrency    *tview.TextView
	runGitMaxTasksInput  *tview.InputField
	runGitCommands       *tview.List
	runGitReadyTasks     *tview.TextView
	selectedBoardID      string
	workflowList         *tview.List
	workflowDetails      *tview.TextView
}

func New(ingest IngestFunc, startTaskTree StartTaskTreeFunc, cancelTaskTree CancelTaskTreeFunc, listTaskboards ListTaskboardsFunc, listReadyTaskIDs ListReadyTaskIDsFunc, listWorkflows ListWorkflowsFunc, getWorkflowStatus WorkflowStatusFunc, copilotAuthStatus CopilotAuthStatusFunc, copilotAuthenticate CopilotAuthenticateFunc, repositoryRoot string, maxTaskLimit int) *UI {
	application := tview.NewApplication().EnableMouse(true)
	status := tview.NewTextView().SetText("Ready").SetDynamicColors(true)
	status.SetBorder(true)
	status.SetTitle("Status")
	if maxTaskLimit < 1 {
		maxTaskLimit = 1
	}

	ui := &UI{
		application:   application,
		pages:         tview.NewPages(),
		status:        status,
		maxTaskLimit:  maxTaskLimit,
		screenStack:   []string{"main"},
		currentScreen: "main",
	}

	mainScreen := ui.buildMainScreen()
	authenticationScreen := ui.buildAuthenticationCommandsScreen(copilotAuthStatus, copilotAuthenticate)
	ingestionScreen := ui.buildIngestionCommandsScreen()
	worktreeScreen := ui.buildWorktreeCommandsScreen()
	runIngestionScreen := ui.buildRunIngestionScreen(ingest)
	runGitflowScreen := ui.buildRunGitflowScreen(startTaskTree, cancelTaskTree, listTaskboards, listReadyTaskIDs, repositoryRoot)
	ingestionWorkflowStatusScreen := ui.buildWorkflowStatusScreen(
		"ingestion_workflows",
		"Ingestion Workflow Status",
		"No ingestion workflows available.",
		listWorkflows,
		getWorkflowStatus,
		func(workflow apptaskboard.IngestionWorkflow) bool {
			return apptaskboard.IsIngestionWorkflowTaskType(workflow.TaskType)
		},
	)
	worktreeWorkflowStatusScreen := ui.buildWorkflowStatusScreen(
		"worktree_workflows",
		"Worktree Workflow Status",
		"No worktree workflows available.",
		listWorkflows,
		getWorkflowStatus,
		func(workflow apptaskboard.IngestionWorkflow) bool {
			return apptaskboard.IsWorktreeWorkflowTaskType(workflow.TaskType)
		},
	)

	ui.pages.AddPage("main", mainScreen, true, true)
	ui.pages.AddPage("authentication_commands", authenticationScreen, true, false)
	ui.pages.AddPage("ingestion_commands", ingestionScreen, true, false)
	ui.pages.AddPage("worktree_commands", worktreeScreen, true, false)
	ui.pages.AddPage("ingestion_run", runIngestionScreen, true, false)
	ui.pages.AddPage("gitflow_run", runGitflowScreen, true, false)
	ui.pages.AddPage("ingestion_workflows", ingestionWorkflowStatusScreen, true, false)
	ui.pages.AddPage("worktree_workflows", worktreeWorkflowStatusScreen, true, false)

	application.SetRoot(ui.pages, true)
	if ui.mainCommandList != nil {
		application.SetFocus(ui.mainCommandList)
	}

	application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			if ui.canGoBack() {
				ui.goBack()
				return nil
			}
		case tcell.KeyTAB:
			ui.toggleFocusForCurrentScreen()
			return nil
		}

		if event.Rune() == 'q' || event.Rune() == 'Q' {
			application.Stop()
			return nil
		}
		return event
	})

	return ui
}

func (ui *UI) Run() error {
	return ui.application.Run()
}

func (ui *UI) Stop() {
	ui.application.Stop()
}

func (ui *UI) buildMainScreen() tview.Primitive {
	header := tview.NewTextView().SetText("Agentic Worktrees Dashboard").SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	header.SetBorder(true)

	body := tview.NewTextView().SetText("Choose a command area: Authentication, Ingestion, or Worktrees.")
	body.SetBorder(true)
	body.SetTitle("Main")

	ui.mainCommandList = ui.newCommandList([]Command{
		{
			Title:       "Authentication",
			Description: "Open Copilot authentication commands",
			Shortcut:    'a',
			Action:      func() { ui.navigateTo("authentication_commands") },
		},
		{
			Title:       "Ingestion",
			Description: "Open ingestion workflow commands",
			Shortcut:    'i',
			Action:      func() { ui.navigateTo("ingestion_commands") },
		},
		{
			Title:       "Worktrees",
			Description: "Open taskboard worktree execution commands",
			Shortcut:    'w',
			Action:      func() { ui.navigateTo("worktree_commands") },
		},
		{
			Title:       "Exit",
			Description: "Close dashboard",
			Shortcut:    'x',
			Action:      func() { ui.application.Stop() },
		},
	}, false)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(body, 3, 0, false).
		AddItem(ui.mainCommandList, 0, 1, true).
		AddItem(ui.status, 3, 0, false)
}

func (ui *UI) buildAuthenticationCommandsScreen(copilotAuthStatus CopilotAuthStatusFunc, copilotAuthenticate CopilotAuthenticateFunc) tview.Primitive {
	header := tview.NewTextView().SetText("Authentication Commands").SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	header.SetBorder(true)

	body := tview.NewTextView().SetText("Select authentication sub-command.")
	body.SetBorder(true)
	body.SetTitle("Authentication")

	ui.authCommandList = ui.newCommandList([]Command{
		{Title: "Copilot Auth Status", Description: "Check current Copilot CLI authentication status", Shortcut: 's', Action: func() {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if copilotAuthStatus == nil {
					ui.application.QueueUpdateDraw(func() { ui.status.SetText("Copilot auth status command unavailable") })
					return
				}
				status, err := copilotAuthStatus(ctx)
				ui.application.QueueUpdateDraw(func() {
					if err != nil {
						ui.status.SetText(fmt.Sprintf("Copilot auth status failed: %s", formatUserError(err)))
						return
					}
					if status == "" {
						status = "Copilot auth status returned no output"
					}
					ui.status.SetText(status)
				})
			}()
		}},
		{Title: "Authenticate Copilot", Description: "Run interactive Copilot login in terminal (tmux)", Shortcut: 'a', Action: func() {
			if copilotAuthenticate == nil {
				ui.status.SetText("Copilot authenticate command unavailable")
				return
			}
			ui.application.Suspend(func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				err := copilotAuthenticate(ctx)
				ui.application.QueueUpdateDraw(func() {
					if err != nil {
						ui.status.SetText(fmt.Sprintf("Copilot authentication failed: %s", formatUserError(err)))
						return
					}
					ui.status.SetText("Copilot authentication completed")
				})
			})
		}},
	}, true)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(body, 3, 0, false).
		AddItem(ui.authCommandList, 0, 1, true).
		AddItem(ui.status, 3, 0, false)
}

func (ui *UI) buildIngestionCommandsScreen() tview.Primitive {
	header := tview.NewTextView().SetText("Ingestion Commands").SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	header.SetBorder(true)

	body := tview.NewTextView().SetText("Select ingestion sub-command.")
	body.SetBorder(true)
	body.SetTitle("Ingestion")

	ui.ingestionCommandList = ui.newCommandList([]Command{
		{Title: "Run Ingestion", Description: "Run new ingestion workflow", Shortcut: 'r', Action: func() { ui.navigateTo("ingestion_run") }},
		{Title: "Workflow Status", Description: "List ingestion workflows and inspect recovery details", Shortcut: 'w', Action: func() { ui.navigateTo("ingestion_workflows") }},
	}, true)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(body, 3, 0, false).
		AddItem(ui.ingestionCommandList, 0, 1, true).
		AddItem(ui.status, 3, 0, false)
}

func (ui *UI) buildWorktreeCommandsScreen() tview.Primitive {
	header := tview.NewTextView().SetText("Worktree Commands").SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	header.SetBorder(true)

	body := tview.NewTextView().SetText("Select worktree sub-command.")
	body.SetBorder(true)
	body.SetTitle("Worktrees")

	ui.worktreeCommandList = ui.newCommandList([]Command{
		{Title: "Run Worktree Pipeline", Description: "Open taskboard execution screen", Shortcut: 'r', Action: func() { ui.navigateTo("gitflow_run") }},
		{Title: "Workflow Status", Description: "List worktree workflows and inspect recovery details", Shortcut: 'w', Action: func() { ui.navigateTo("worktree_workflows") }},
	}, true)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(body, 3, 0, false).
		AddItem(ui.worktreeCommandList, 0, 1, true).
		AddItem(ui.status, 3, 0, false)
}

func (ui *UI) buildRunGitflowScreen(startTaskTree StartTaskTreeFunc, cancelTaskTree CancelTaskTreeFunc, listTaskboards ListTaskboardsFunc, listReadyTaskIDs ListReadyTaskIDsFunc, repositoryRoot string) tview.Primitive {
	header := tview.NewTextView().SetText("Start Task Tree Pipeline").SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	header.SetBorder(true)

	repositoryRootLabel := tview.NewTextView().SetDynamicColors(true)
	repositoryRootLabel.SetBorder(true)
	repositoryRootLabel.SetTitle("Repository Root")
	repositoryRootLabel.SetText(repositoryRoot)

	boardList := tview.NewList().ShowSecondaryText(true)
	boardList.SetBorder(true)
	boardList.SetTitle("Taskboards")
	ui.runGitBoardList = boardList

	readyTasks := tview.NewTextView().SetDynamicColors(true)
	readyTasks.SetBorder(true)
	readyTasks.SetTitle("Ready Tasks")
	readyTasks.SetText("Select a taskboard and refresh to load ready tasks.")
	ui.runGitReadyTasks = readyTasks

	sourceBranchInput := tview.NewInputField().SetLabel("Source Branch: ")
	sourceBranchInput.SetBorder(true)
	sourceBranchInput.SetTitle("Source Branch")
	ui.runGitSourceInput = sourceBranchInput

	concurrencyView := tview.NewTextView().SetDynamicColors(true)
	concurrencyView.SetBorder(true)
	concurrencyView.SetTitle("Concurrent Agents")
	concurrencyView.SetText(fmt.Sprintf("%d (derived from worker concurrency)", ui.maxTaskLimit))
	ui.runGitConcurrency = concurrencyView

	maxTasksInput := tview.NewInputField().SetLabel("Max Tasks (optional): ")
	maxTasksInput.SetBorder(true)
	maxTasksInput.SetTitle("Task Limit (total tasks to execute)")
	ui.runGitMaxTasksInput = maxTasksInput

	loadReadyTasks := func(boardID string) {
		cleanBoardID := strings.TrimSpace(boardID)
		if cleanBoardID == "" {
			return
		}
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			if listReadyTaskIDs == nil {
				ui.application.QueueUpdateDraw(func() { ui.status.SetText("Ready task query unavailable") })
				return
			}
			readyTaskIDs, err := listReadyTaskIDs(ctx, cleanBoardID)
			ui.application.QueueUpdateDraw(func() {
				if err != nil {
					ui.status.SetText(fmt.Sprintf("Load ready tasks failed: %s", formatUserError(err)))
					return
				}
				if len(readyTaskIDs) == 0 {
					readyTasks.SetText("No ready tasks for selected board.")
					return
				}
				readyTasks.SetText(strings.Join(readyTaskIDs, "\n"))
			})
		}()
	}

	refreshBoards := func() {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			if listTaskboards == nil {
				ui.application.QueueUpdateDraw(func() { ui.status.SetText("Taskboard list unavailable") })
				return
			}
			boardIDs, err := listTaskboards(ctx)
			ui.application.QueueUpdateDraw(func() {
				if err != nil {
					ui.status.SetText(fmt.Sprintf("Load taskboards failed: %s", formatUserError(err)))
					return
				}
				boardList.Clear()
				if len(boardIDs) == 0 {
					ui.selectedBoardID = ""
					boardList.AddItem("(none)", "No taskboards found", 0, nil)
					readyTasks.SetText("No taskboards available.")
					ui.status.SetText("No taskboards found")
					return
				}
				for _, boardID := range boardIDs {
					currentBoardID := boardID
					boardList.AddItem(currentBoardID, "Select board", 0, func() {
						ui.selectedBoardID = currentBoardID
						ui.status.SetText(fmt.Sprintf("Selected board %s", currentBoardID))
						loadReadyTasks(currentBoardID)
					})
				}
				if ui.selectedBoardID == "" {
					ui.selectedBoardID = boardIDs[0]
					loadReadyTasks(ui.selectedBoardID)
				}
				ui.status.SetText(fmt.Sprintf("Loaded %d taskboards", len(boardIDs)))
			})
		}()
	}

	ui.runGitCommands = ui.newCommandList([]Command{
		{
			Title:       "Refresh Taskboards",
			Description: "Reload all taskboards from local JSON storage",
			Shortcut:    'f',
			Action:      refreshBoards,
		},
		{
			Title:       "Cancel Runner",
			Description: "Cancel and cleanup running taskboard runner for selected board",
			Shortcut:    'x',
			Action: func() {
				boardID := strings.TrimSpace(ui.selectedBoardID)
				if boardID == "" {
					ui.status.SetText("Select a taskboard first")
					return
				}
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
					defer cancel()
					if cancelTaskTree == nil {
						ui.application.QueueUpdateDraw(func() { ui.status.SetText("Task tree cancel command unavailable") })
						return
					}
					result, err := cancelTaskTree(ctx, boardID)
					ui.application.QueueUpdateDraw(func() {
						if err != nil {
							ui.status.SetText(fmt.Sprintf("Task tree cancel failed: %s", formatUserError(err)))
							return
						}
						ui.status.SetText(result)
					})
				}()
			},
		},
		{
			Title:       "Execute",
			Description: "Start task tree pipeline for selected taskboard",
			Shortcut:    'e',
			Action: func() {
				boardID := strings.TrimSpace(ui.selectedBoardID)
				if boardID == "" {
					ui.status.SetText("Select a taskboard first")
					return
				}
				sourceBranch := strings.TrimSpace(sourceBranchInput.GetText())
				if sourceBranch == "" {
					ui.status.SetText("Source branch is required")
					return
				}
				maxTasks := 0
				maxTasksText := strings.TrimSpace(maxTasksInput.GetText())
				if maxTasksText != "" {
					parsedMaxTasks, parseErr := strconv.Atoi(maxTasksText)
					if parseErr != nil || parsedMaxTasks <= 0 {
						ui.status.SetText("Max Tasks must be a positive integer (or leave blank)")
						return
					}
					maxTasks = parsedMaxTasks
				}
				ui.status.SetText(fmt.Sprintf("Loading ready tasks for board %s ...", boardID))
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
					defer cancel()
					if listReadyTaskIDs == nil {
						ui.application.QueueUpdateDraw(func() { ui.status.SetText("Ready task query unavailable") })
						return
					}
					readyTaskIDs, err := listReadyTaskIDs(ctx, boardID)
					if err != nil {
						ui.application.QueueUpdateDraw(func() {
							ui.status.SetText(fmt.Sprintf("Load ready tasks failed: %s", formatUserError(err)))
						})
						return
					}
					if len(readyTaskIDs) == 0 {
						ui.application.QueueUpdateDraw(func() { ui.status.SetText("No ready tasks in selected taskboard") })
						return
					}
					if startTaskTree == nil {
						ui.application.QueueUpdateDraw(func() { ui.status.SetText("Task tree start command unavailable") })
						return
					}
					queueTaskID, err := startTaskTree(ctx, boardID, sourceBranch, maxTasks)
					ui.application.QueueUpdateDraw(func() {
						if err != nil {
							ui.status.SetText(fmt.Sprintf("Task tree start failed: %s", formatUserError(err)))
							return
						}
						if maxTasks > 0 {
							ui.status.SetText(fmt.Sprintf("Started task tree for board %s (queue task %s, max tasks %d)", boardID, queueTaskID, maxTasks))
							return
						}
						ui.status.SetText(fmt.Sprintf("Started task tree for board %s (queue task %s)", boardID, queueTaskID))
					})
				}()
			},
		},
	}, true)

	content := tview.NewFlex().
		AddItem(boardList, 0, 1, true).
		AddItem(readyTasks, 0, 1, false)

	refreshBoards()

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(repositoryRootLabel, 3, 0, false).
		AddItem(content, 0, 1, true).
		AddItem(sourceBranchInput, 3, 0, false).
		AddItem(concurrencyView, 3, 0, false).
		AddItem(maxTasksInput, 3, 0, false).
		AddItem(ui.runGitCommands, 0, 1, false).
		AddItem(ui.status, 3, 0, false)
}

func (ui *UI) buildRunIngestionScreen(ingest IngestFunc) tview.Primitive {
	header := tview.NewTextView().SetText("Run Ingestion").SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	header.SetBorder(true)

	input := tview.NewInputField().SetLabel("Directory: ")
	input.SetBorder(true)
	input.SetTitle("Ingest Source")
	ui.runIngestionInput = input

	ui.runIngestionCommands = ui.newCommandList([]Command{
		{
			Title:       "Execute",
			Description: "Start ingestion for provided directory",
			Shortcut:    'e',
			Action: func() {
				directory := strings.TrimSpace(input.GetText())
				if directory == "" {
					ui.status.SetText("Directory is required")
					return
				}
				ui.status.SetText(fmt.Sprintf("Ingesting %s ...", directory))
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
					defer cancel()
					if ingest == nil {
						ui.application.QueueUpdateDraw(func() { ui.status.SetText("Ingest command unavailable") })
						return
					}
					result, err := ingest(ctx, directory)
					ui.application.QueueUpdateDraw(func() {
						if err != nil {
							ui.status.SetText(fmt.Sprintf("Ingest failed: %s", formatUserError(err)))
							return
						}
						ui.status.SetText(fmt.Sprintf("Created board %s (run %s)", result.BoardID, result.RunID))
					})
				}()
			},
		},
	}, true)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(input, 3, 0, true).
		AddItem(ui.runIngestionCommands, 0, 1, false).
		AddItem(ui.status, 3, 0, false)
}

func (ui *UI) buildWorkflowStatusScreen(screenID string, headerTitle string, emptyMessage string, listWorkflows ListWorkflowsFunc, getWorkflowStatus WorkflowStatusFunc, includeWorkflow func(workflow apptaskboard.IngestionWorkflow) bool) tview.Primitive {
	header := tview.NewTextView().SetText(strings.TrimSpace(headerTitle)).SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	header.SetBorder(true)

	workflowList := tview.NewList().ShowSecondaryText(true)
	workflowList.SetBorder(true)
	workflowList.SetTitle("Workflows")
	ui.workflowList = workflowList

	details := tview.NewTextView().SetDynamicColors(true)
	details.SetBorder(true)
	details.SetTitle("Workflow Details")
	details.SetText("Select a workflow.")
	ui.workflowDetails = details

	loadWorkflowDetails := func(runID string) {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if getWorkflowStatus == nil {
				ui.application.QueueUpdateDraw(func() { ui.status.SetText("Workflow details unavailable") })
				return
			}
			workflow, err := getWorkflowStatus(ctx, runID)
			ui.application.QueueUpdateDraw(func() {
				if err != nil {
					ui.status.SetText(fmt.Sprintf("Workflow lookup failed: %s", formatUserError(err)))
					details.SetText("Workflow no longer exists in Asynq. Refresh to reload active statuses.")
					return
				}
				stream := workflow.Stream
				if strings.TrimSpace(stream) == "" {
					stream = "(no stream details recorded yet)"
				}
				details.SetText(fmt.Sprintf("RunID: %s\nType: %s\nStatus: %s\nTaskID: %s\nBoardID: %s\nMessage: %s\nUpdated: %s\n\nStream:\n%s", workflow.RunID, workflow.TaskType, workflow.Status, workflow.TaskID, workflow.BoardID, workflow.Message, workflow.UpdatedAt.Format(time.RFC3339), stream))
				ui.status.SetText(fmt.Sprintf("Loaded workflow %s", workflow.RunID))
			})
		}()
	}

	refreshWorkflows := func() {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if listWorkflows == nil {
				ui.application.QueueUpdateDraw(func() { ui.status.SetText("Workflow list unavailable") })
				return
			}
			workflows, err := listWorkflows(ctx)
			ui.application.QueueUpdateDraw(func() {
				if err != nil {
					ui.status.SetText(fmt.Sprintf("Workflow list failed: %s", formatUserError(err)))
					return
				}
				workflowList.Clear()
				filteredWorkflows := make([]apptaskboard.IngestionWorkflow, 0, len(workflows))
				for _, workflow := range workflows {
					if includeWorkflow != nil && !includeWorkflow(workflow) {
						continue
					}
					filteredWorkflows = append(filteredWorkflows, workflow)
				}
				if len(filteredWorkflows) == 0 {
					workflowList.AddItem("(none)", strings.TrimSpace(emptyMessage), 0, nil)
					details.SetText(strings.TrimSpace(emptyMessage))
					ui.status.SetText(strings.TrimSpace(emptyMessage))
					return
				}
				for _, workflow := range filteredWorkflows {
					runID := workflow.RunID
					statusText := string(workflow.Status)
					if strings.TrimSpace(workflow.TaskType) != "" {
						statusText = fmt.Sprintf("%s | %s", statusText, strings.TrimSpace(workflow.TaskType))
					}
					workflowList.AddItem(runID, statusText, 0, func() {
						loadWorkflowDetails(runID)
					})
				}
				ui.status.SetText(fmt.Sprintf("Loaded %d workflows", len(filteredWorkflows)))
			})
		}()
	}

	commands := ui.newCommandList([]Command{
		{
			Title:       "Back",
			Description: "Return to previous screen",
			Shortcut:    'b',
			Action:      func() { ui.goBack() },
		},
		{
			Title:       "Refresh Workflows",
			Description: "Reload workflow statuses from Asynq",
			Shortcut:    'f',
			Action:      refreshWorkflows,
		},
	}, false)

	content := tview.NewFlex().
		AddItem(workflowList, 0, 1, true).
		AddItem(details, 0, 2, false)

	screen := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(commands, 6, 0, false).
		AddItem(content, 0, 1, true).
		AddItem(ui.status, 3, 0, false)

	refreshWorkflows()
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if ui.currentScreen == strings.TrimSpace(screenID) {
				refreshWorkflows()
			}
		}
	}()
	return screen
}

func (ui *UI) newCommandList(commands []Command, includeBack bool) *tview.List {
	commandList := tview.NewList().ShowSecondaryText(true)
	commandList.SetBorder(true)
	commandList.SetTitle("Commands")

	for _, command := range commands {
		commandAction := command.Action
		commandList.AddItem(command.Title, command.Description, command.Shortcut, commandAction)
	}

	if includeBack {
		commandList.AddItem("Back", "Return to previous screen", 'b', func() { ui.goBack() })
	}

	return commandList
}

func (ui *UI) navigateTo(screenID string) {
	if len(ui.screenStack) == 0 || ui.screenStack[len(ui.screenStack)-1] != screenID {
		ui.screenStack = append(ui.screenStack, screenID)
	}
	ui.currentScreen = screenID
	ui.pages.SwitchToPage(screenID)
	ui.focusCurrentScreenDefault()
}

func (ui *UI) canGoBack() bool {
	return len(ui.screenStack) > 1
}

func (ui *UI) goBack() {
	if !ui.canGoBack() {
		return
	}
	ui.screenStack = ui.screenStack[:len(ui.screenStack)-1]
	previousScreen := ui.screenStack[len(ui.screenStack)-1]
	ui.currentScreen = previousScreen
	ui.pages.SwitchToPage(previousScreen)
	ui.focusCurrentScreenDefault()
}

func (ui *UI) focusCurrentScreenDefault() {
	switch ui.currentScreen {
	case "authentication_commands":
		if ui.authCommandList != nil {
			ui.application.SetFocus(ui.authCommandList)
		}
	case "ingestion_commands":
		if ui.ingestionCommandList != nil {
			ui.application.SetFocus(ui.ingestionCommandList)
		}
	case "worktree_commands":
		if ui.worktreeCommandList != nil {
			ui.application.SetFocus(ui.worktreeCommandList)
		}
	case "ingestion_run":
		if ui.runIngestionInput != nil {
			ui.application.SetFocus(ui.runIngestionInput)
		}
	case "ingestion_workflows":
		if ui.workflowList != nil {
			ui.application.SetFocus(ui.workflowList)
		}
	case "gitflow_run":
		if ui.runGitBoardList != nil {
			ui.application.SetFocus(ui.runGitBoardList)
		}
	default:
		if ui.mainCommandList != nil {
			ui.application.SetFocus(ui.mainCommandList)
		}
	}
}

func (ui *UI) toggleFocusForCurrentScreen() {
	switch ui.currentScreen {
	case "ingestion_run":
		if ui.runIngestionInput == nil || ui.runIngestionCommands == nil {
			return
		}
		if ui.application.GetFocus() == ui.runIngestionInput {
			ui.application.SetFocus(ui.runIngestionCommands)
			return
		}
		ui.application.SetFocus(ui.runIngestionInput)
	case "ingestion_workflows":
		if ui.workflowList == nil || ui.workflowDetails == nil {
			return
		}
		if ui.application.GetFocus() == ui.workflowList {
			ui.application.SetFocus(ui.workflowDetails)
			return
		}
		ui.application.SetFocus(ui.workflowList)
	case "gitflow_run":
		if ui.runGitBoardList == nil || ui.runGitSourceInput == nil || ui.runGitConcurrency == nil || ui.runGitMaxTasksInput == nil || ui.runGitCommands == nil {
			return
		}
		if ui.application.GetFocus() == ui.runGitBoardList {
			ui.application.SetFocus(ui.runGitSourceInput)
			return
		}
		if ui.application.GetFocus() == ui.runGitSourceInput {
			ui.application.SetFocus(ui.runGitMaxTasksInput)
			return
		}
		if ui.application.GetFocus() == ui.runGitMaxTasksInput {
			ui.application.SetFocus(ui.runGitCommands)
			return
		}
		ui.application.SetFocus(ui.runGitBoardList)
	}
}

func formatUserError(err error) string {
	if err == nil {
		return "unknown error"
	}
	message := strings.TrimSpace(err.Error())
	lower := strings.ToLower(message)
	if strings.Contains(lower, "start copilot client") {
		return "Copilot could not start. Check GitHub auth, Copilot access, and CLI/token settings."
	}
	if strings.Contains(lower, "workflow not found") {
		return "Workflow not found in Asynq. Refresh the list and try again."
	}
	if strings.Contains(lower, "context deadline exceeded") {
		return "Request timed out. Try again and inspect the log for detailed progress."
	}
	return message
}
