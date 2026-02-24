package dashboard

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	application *tview.Application
}

func New() *UI {
	application := tview.NewApplication()

	header := tview.NewTextView().
		SetText("Agentic Worktrees Dashboard").
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	header.SetBorder(true)

	status := tview.NewTextView().
		SetText("Ready").
		SetDynamicColors(true)
	status.SetBorder(true)
	status.SetTitle("Status")

	commands := tview.NewList().
		ShowSecondaryText(true)
	commands.SetBorder(true)
	commands.SetTitle("Commands")
	commands.AddItem(
		"Ingest Documentation",
		"Start ingestion flow for documentation-to-task decomposition",
		'i',
		func() {
			status.SetText("Ingest Documentation selected (initial placeholder)")
		},
	)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(commands, 0, 1, true).
		AddItem(status, 3, 0, false)

	application.SetRoot(layout, true)
	application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' || event.Rune() == 'Q' {
			application.Stop()
			return nil
		}
		return event
	})

	return &UI{application: application}
}

func (ui *UI) Run() error {
	return ui.application.Run()
}

func (ui *UI) Stop() {
	ui.application.Stop()
}
