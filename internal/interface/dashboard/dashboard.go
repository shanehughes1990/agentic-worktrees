package dashboard

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	apptaskboard "github.com/shanehughes1990/agentic-worktrees/internal/application/taskboard"
)

type IngestFunc func(ctx context.Context, directory string) (apptaskboard.IngestionResult, error)
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
	screenStack          []string
	currentScreen        string
	mainCommandList      *tview.List
	ingestionCommandList *tview.List
	runIngestionInput    *tview.InputField
	runIngestionCommands *tview.List
	workflowList         *tview.List
	workflowDetails      *tview.TextView
}

func New(ingest IngestFunc, listWorkflows ListWorkflowsFunc, getWorkflowStatus WorkflowStatusFunc, copilotAuthStatus CopilotAuthStatusFunc, copilotAuthenticate CopilotAuthenticateFunc) *UI {
	application := tview.NewApplication().EnableMouse(true)
	status := tview.NewTextView().SetText("Ready").SetDynamicColors(true)
	status.SetBorder(true)
	status.SetTitle("Status")

	ui := &UI{
		application:   application,
		pages:         tview.NewPages(),
		status:        status,
		screenStack:   []string{"main"},
		currentScreen: "main",
	}

	mainScreen := ui.buildMainScreen()
	ingestionScreen := ui.buildIngestionCommandsScreen(copilotAuthStatus, copilotAuthenticate)
	runIngestionScreen := ui.buildRunIngestionScreen(ingest)
	workflowStatusScreen := ui.buildWorkflowStatusScreen(listWorkflows, getWorkflowStatus)

	ui.pages.AddPage("main", mainScreen, true, true)
	ui.pages.AddPage("ingestion_commands", ingestionScreen, true, false)
	ui.pages.AddPage("ingestion_run", runIngestionScreen, true, false)
	ui.pages.AddPage("ingestion_workflows", workflowStatusScreen, true, false)

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

	body := tview.NewTextView().SetText("Choose a command area.")
	body.SetBorder(true)
	body.SetTitle("Main")

	ui.mainCommandList = ui.newCommandList([]Command{{
		Title:       "Ingestion",
		Description: "Open ingestion command area",
		Shortcut:    'i',
		Action:      func() { ui.navigateTo("ingestion_commands") },
	}}, false)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(body, 3, 0, false).
		AddItem(ui.mainCommandList, 0, 1, true).
		AddItem(ui.status, 3, 0, false)
}

func (ui *UI) buildIngestionCommandsScreen(copilotAuthStatus CopilotAuthStatusFunc, copilotAuthenticate CopilotAuthenticateFunc) tview.Primitive {
	header := tview.NewTextView().SetText("Ingestion Commands").SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	header.SetBorder(true)

	body := tview.NewTextView().SetText("Select ingestion sub-command.")
	body.SetBorder(true)
	body.SetTitle("Ingestion")

	ui.ingestionCommandList = ui.newCommandList([]Command{
		{Title: "Run Ingestion", Description: "Run new ingestion workflow", Shortcut: 'r', Action: func() { ui.navigateTo("ingestion_run") }},
		{Title: "Workflow Status", Description: "List live Asynq workflows and inspect details", Shortcut: 'w', Action: func() { ui.navigateTo("ingestion_workflows") }},
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
		AddItem(ui.ingestionCommandList, 0, 1, true).
		AddItem(ui.status, 3, 0, false)
}

func (ui *UI) buildRunIngestionScreen(ingest IngestFunc) tview.Primitive {
	header := tview.NewTextView().SetText("Run Ingestion").SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	header.SetBorder(true)

	input := tview.NewInputField().SetLabel("Directory: ")
	input.SetBorder(true)
	input.SetTitle("Ingest Source")
	ui.runIngestionInput = input

	ui.runIngestionCommands = ui.newCommandList([]Command{{
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
	}}, true)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(input, 3, 0, true).
		AddItem(ui.runIngestionCommands, 0, 1, false).
		AddItem(ui.status, 3, 0, false)
}

func (ui *UI) buildWorkflowStatusScreen(listWorkflows ListWorkflowsFunc, getWorkflowStatus WorkflowStatusFunc) tview.Primitive {
	header := tview.NewTextView().SetText("Live Asynq Ingestion Workflow Status").SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
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
				details.SetText(fmt.Sprintf("RunID: %s\nStatus: %s\nTaskID: %s\nBoardID: %s\nMessage: %s\nUpdated: %s\n\nStream:\n%s", workflow.RunID, workflow.Status, workflow.TaskID, workflow.BoardID, workflow.Message, workflow.UpdatedAt.Format(time.RFC3339), stream))
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
				if len(workflows) == 0 {
					workflowList.AddItem("(none)", "No live Asynq ingestion workflows found", 0, nil)
					details.SetText("No live Asynq workflows available.")
					ui.status.SetText("No live Asynq workflows found")
					return
				}
				for _, workflow := range workflows {
					runID := workflow.RunID
					statusText := string(workflow.Status)
					workflowList.AddItem(runID, statusText, 0, func() {
						loadWorkflowDetails(runID)
					})
				}
				ui.status.SetText(fmt.Sprintf("Loaded %d live Asynq workflows", len(workflows)))
			})
		}()
	}

	commands := ui.newCommandList([]Command{{
		Title:       "Refresh Workflows",
		Description: "Reload all live ingestion workflow statuses from Asynq",
		Shortcut:    'f',
		Action:      refreshWorkflows,
	}}, true)

	content := tview.NewFlex().
		AddItem(workflowList, 0, 1, true).
		AddItem(details, 0, 2, false)

	screen := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(commands, 4, 0, false).
		AddItem(content, 0, 1, true).
		AddItem(ui.status, 3, 0, false)

	refreshWorkflows()
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if ui.currentScreen == "ingestion_workflows" {
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
	case "ingestion_commands":
		if ui.ingestionCommandList != nil {
			ui.application.SetFocus(ui.ingestionCommandList)
		}
	case "ingestion_run":
		if ui.runIngestionInput != nil {
			ui.application.SetFocus(ui.runIngestionInput)
		}
	case "ingestion_workflows":
		if ui.workflowList != nil {
			ui.application.SetFocus(ui.workflowList)
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
