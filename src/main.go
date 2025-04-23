package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	// Load configuration
	cfg, err := LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Cannot load config: %v", err)
	}

	// Start background process watchers
	procEvents, cancel := StartProcessWatchers(cfg.Processes, len(cfg.Processes)*2)
	defer cancel()

	// Create the application
	app := tview.NewApplication()

	// UI components
	tabBar := NewTabBar()
	footer := NewFooter()
    UpdateFooter(footer, "*Main (ctrl+F1)", "Logs (ctrl+F2)", "Procs (ctrl+F3)")	

	// ── Main view ────────────────────────────────────────────────────────────────
	splashPages := tview.NewPages()
	output := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	output.SetBorder(true).SetTitle(" Output ")

	var rootMenu tview.Primitive
	menuList := BuildMenu(app, splashPages, cfg.Menu, output, &rootMenu, true)
	splashPages.AddAndSwitchToPage("main", menuList, true)

	input := tview.NewInputField().
		SetLabel("> ").
		SetFieldWidth(0)
	input.SetBorder(true).SetTitle(" Shell Input ")
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			cmd := input.GetText()
			input.SetText("")
			executeCommand(cmd, output)
			app.SetFocus(rootMenu)
		}
	})

	mainView := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().
				AddItem(splashPages, 0, 1, true).
				AddItem(output, 0, 2, false),
			0, 1, true,
		).
		AddItem(input, 3, 0, false)

	// ── Tail view ────────────────────────────────────────────────────────────────
	tailView := BuildTailView(app, cfg.TailFiles)

	// ── Process view ─────────────────────────────────────────────────────────────
	procList, procView := NewProcessPanel(app, cfg.Processes, procEvents)

	// Pages container
	pages := tview.NewPages().
		AddPage("main", mainView, true, true).
		AddPage("tails", tailView, true, false).
		AddPage("proc", procView, true, false)

	// Root layout: tab bar, pages, footer
	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tabBar, 1, 0, false).
		AddItem(pages, 0, 1, true).
		AddItem(footer, 3, 0, false)

	// Tab switching (F1/F2/F3)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			pages.SwitchToPage("main")
			UpdateFooter(footer, "*Main (ctrl+F1)", "Logs (ctrl+F2)", "Procs (ctrl+F3)")
			app.SetFocus(rootMenu)
			return nil
		case tcell.KeyF2:
			pages.SwitchToPage("tails")
			UpdateFooter(footer, "Main (ctrl+F1) ", "*Logs (ctrl+F2)", "Procs (ctrl+F3)")
			return nil
		case tcell.KeyF3:
			pages.SwitchToPage("proc")
			UpdateFooter(footer, "Main (ctrl+F1) ", "Logs (ctrl+F2)", "*Procs (ctrl+F3)")
			app.SetFocus(procList)
			return nil
		}
		return event
	})

	// Run application
	if err := app.SetRoot(root, true).EnableMouse(true).Run(); err != nil {
		log.Fatal(err)
	}
}
