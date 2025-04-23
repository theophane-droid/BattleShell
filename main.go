package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/theophane-droid/battleshell/battleshell"
)

func main() {
	app := tview.NewApplication()

	// ── UI placeholders ───────────────────────────────────────
	tabBar := battleshell.NewTabBar()
	footer := battleshell.NewFooter()

	var (
		rootMenu        tview.Primitive
		menuPages       *tview.Pages
		tailView        tview.Primitive
		procList        *tview.List
		procView        tview.Primitive
		mainView        tview.Primitive
		pages           *tview.Pages
		root            *tview.Flex
		procEvents      <-chan battleshell.ProcEvent
		cancelWatchers  func()
		reloadEverything func() // declared, body added later
	)

	// ── reloadEverything definition ───────────────────────────
	reloadEverything = func() {
		if cancelWatchers != nil {
			cancelWatchers()
		}

		cfg, err := battleshell.LoadConfig("config.json")
		if err != nil {
			modal := tview.NewModal().
				SetText("Invalid config.json:\n\n" + err.Error()).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(int, string) { app.SetRoot(root, true) })
			app.SetRoot(modal, true)
			return
		}

		procEvents, cancelWatchers = battleshell.StartProcessWatchers(cfg.Processes, len(cfg.Processes)*2)

		// Build menu / tail / proc
		menuPages = tview.NewPages()
		output := tview.NewTextView()
		output.SetDynamicColors(true).
			SetScrollable(true).
			SetWrap(true).
			SetBorder(true).SetTitle(" Output ")

		menu := battleshell.BuildMenu(app, menuPages, cfg.Menu, output, &rootMenu, true)
		menuPages.AddPage("main", menu, true, true)

		input := tview.NewInputField()
		input.SetLabel("> ").SetFieldWidth(0).
			SetBorder(true).SetTitle(" Shell Input ")
		input.SetDoneFunc(func(k tcell.Key) {
			if k == tcell.KeyEnter {
				cmd := input.GetText()
				input.SetText("")
				battleshell.ExecuteCommand(cmd, output)
				app.SetFocus(rootMenu)
			}
		})

		mainView = tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().
				AddItem(menuPages, 0, 1, true).
				AddItem(output, 0, 2, false), 0, 1, true).
			AddItem(input, 3, 0, false)

		tailView = battleshell.BuildTailView(app, cfg.TailFiles)
		procList, procView = battleshell.NewProcessPanel(app, cfg.Processes, procEvents)

		// Config editor needs reloadEverything, so create after it exists
		confPanel := battleshell.NewConfigEditor(app, "config.json", reloadEverything)

		// Re-compose pages
		pages = tview.NewPages().
			AddPage("main", mainView, true, true).
			AddPage("tails", tailView, true, false).
			AddPage("proc", procView, true, false).
			AddPage("conf", confPanel, true, false)

		// Root flex
		if root == nil {
			root = tview.NewFlex().SetDirection(tview.FlexRow)
		}
		root.Clear()
		root.
			AddItem(tabBar, 1, 0, false).
			AddItem(pages, 0, 1, true).
			AddItem(footer, 3, 0, false)

		app.SetRoot(root, true)
		battleshell.UpdateFooter(footer, "*Main (F1)", "Logs (F2)", "Procs (F3)", "Conf (F4)")
	}

	// First load
	reloadEverything()
	battleshell.UpdateFooter(footer, "*Main (F1)", "Logs (F2)", "Procs (F3)", "Conf (F4)")

	// ── Global shortcuts ──────────────────────────────────────
	app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyF1:
			pages.SwitchToPage("main")
			battleshell.UpdateFooter(footer, "*Main (F1)", "Logs (F2)", "Procs (F3)", "Conf (F4)")
			app.SetFocus(rootMenu)
			return nil
		case tcell.KeyF2:
			pages.SwitchToPage("tails")
			battleshell.UpdateFooter(footer, "Main (F1)", "*Logs (F2)", "Procs (F3)", "Conf (F4)")
			return nil
		case tcell.KeyF3:
			pages.SwitchToPage("proc")
			battleshell.UpdateFooter(footer, "Main (F1)", "Logs (F2)", "*Procs (F3)", "Conf (F4)")
			app.SetFocus(procList)
			return nil
		case tcell.KeyF4:
			pages.SwitchToPage("conf")
			battleshell.UpdateFooter(footer, "Main (F1)", "Logs (F2)", "Procs (F3)", "*Conf (F4)")
			return nil
		}
		return ev
	})

	if err := app.SetRoot(root, true).EnableMouse(true).Run(); err != nil {
		log.Fatal(err)
	}
}
