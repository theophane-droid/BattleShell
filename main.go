package main

import (
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/theophane-droid/battleshell/battleshell"
)

var inForm bool

func main() {
	app := tview.NewApplication()

	/* ────────── static widgets ────────── */
	tabBar := battleshell.NewTabBar()
	footer := battleshell.NewFooter()

	/* ────────── shared vars ───────────── */
	var (
		rootMenu        tview.Primitive
		menuPages       *tview.Pages
		tailView        tview.Primitive
		procView        tview.Primitive
		procList        tview.Primitive
		confPanel       tview.Primitive
		mainView        tview.Primitive
		pages           *tview.Pages
		root            *tview.Flex
		focusables      []tview.Primitive
		curFocus        int
		procEvents      <-chan battleshell.ProcEvent
		cancelWatchers  func()
		reloadEverything func()
		// zoom state
		zoomed          bool
		origRoot        tview.Primitive
		currentPage     string
	)

	/* ────────── persistent widgets ────── */
	output := tview.NewTextArea()
	output.SetBorder(true).
		SetTitle(" Output ")
	output.SetText("", true)

	input := tview.NewInputField()
	input.SetLabel("> ").
		SetFieldWidth(0).
		SetBorder(true).
		SetTitle(" Shell Input ")
	input.SetDoneFunc(func(k tcell.Key) {
		if k == tcell.KeyEnter {
			cmd := input.GetText()
			input.SetText("")
			battleshell.ExecuteCommand(cmd, output, app)
		}
	})

	/* ────────── reloadEverything ──────── */
	reloadEverything = func() {
		if cancelWatchers != nil {
			cancelWatchers()
		}

		cfg, err := battleshell.LoadConfig("config.json")
		if err != nil {
			inForm = true
			modal := tview.NewModal().
				SetText("Invalid config.json:\n\n" + err.Error()).
				AddButtons([]string{"OK"}).
				SetDoneFunc(func(_ int, _ string) {
					inForm = false
					app.Stop()
					os.Exit(1)
				}).
				SetFocus(0)

			app.SetRoot(modal, true)
			return
		}

		procEvents, cancelWatchers = battleshell.StartProcessWatchers(cfg.Processes, len(cfg.Processes)*2)

		/* build MAIN page */
		rootMenu = nil
		menuPages = tview.NewPages()
		mainMenu := battleshell.BuildMenu(app, menuPages, cfg.Menu, output, &rootMenu, true, &inForm)
		menuPages.AddPage("main", mainMenu, true, true)

		mainView = tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(
				tview.NewFlex().
					AddItem(menuPages, 0, 1, true).
					AddItem(output,    0, 2, false),
				0, 1, true).
			AddItem(input, 3, 0, false)

		/* rebuild other pages */
		tailView = battleshell.BuildTailView(app, cfg.TailFiles)
		procList, procView = battleshell.NewProcessPanel(app, cfg.Processes, procEvents)
		confPanel = battleshell.NewConfigEditor(app, "config.json", reloadEverything)

		pages = tview.NewPages().
			AddPage("main", mainView, true,  true).
			AddPage("tails", tailView, true,  false).
			AddPage("proc",  procView, true,  false).
			AddPage("conf",  confPanel, true, false)

		if root == nil {
			root = tview.NewFlex().SetDirection(tview.FlexRow)
		}
		root.Clear().
			AddItem(tabBar, 1, 0, false).
			AddItem(pages,  0, 1, true).
			AddItem(footer, 3, 0, false)

		// save original layout
		if origRoot == nil {
			origRoot = root
		}
		currentPage = "main"

		focusables = []tview.Primitive{menuPages, output, input}
		curFocus = 0

		app.SetRoot(root, true)
		app.SetFocus(focusables[curFocus])
		battleshell.UpdateFooter(footer, "*Main (F1)", "Logs (F2)", "Procs (F3)", "Conf (F4)")
	}

	/* first load */
	reloadEverything()

	/* ────────── Global keys ───────────── */
	app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		// toggle zoom on Ctrl+Z
		if ev.Key() == tcell.KeyCtrlZ {
			if zoomed {
				app.SetRoot(origRoot, true)
				zoomed = false
				// restore focus
				app.SetFocus(focusables[curFocus])
			} else {
				var panel tview.Primitive
				switch currentPage {
				case "main":
					panel = focusables[curFocus]
				case "tails":
					panel = tailView
				case "proc":
					panel = procView
				case "conf":
					panel = confPanel
				}
				zoomView := tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(panel, 0, 1, true).
					AddItem(footer, 3, 0, false)
				app.SetRoot(zoomView, true)
				zoomed = true
				app.SetFocus(panel)
			}
			return nil
		}

		if inForm {
			return ev
		}

		switch ev.Key() {
		case tcell.KeyTab:
			curFocus = (curFocus + 1) % len(focusables)
			app.SetFocus(focusables[curFocus])
			return nil
		case tcell.KeyBacktab:
			curFocus = (curFocus - 1 + len(focusables)) % len(focusables)
			app.SetFocus(focusables[curFocus])
			return nil
		case tcell.KeyF1:
			pages.SwitchToPage("main")
			battleshell.UpdateFooter(footer, "*Main (F1)", "Logs (F2)", "Procs (F3)", "Conf (F4)")
			focusables = []tview.Primitive{menuPages, output, input}
			curFocus = 0
			app.SetFocus(menuPages)
			currentPage = "main"
			return nil
		case tcell.KeyF2:
			pages.SwitchToPage("tails")
			battleshell.UpdateFooter(footer, "Main (F1)", "*Logs (F2)", "Procs (F3)", "Conf (F4)")
			if flex, ok := tailView.(*tview.Flex); ok && flex.GetItemCount() >= 2 {
				focusables = []tview.Primitive{flex.GetItem(0), flex.GetItem(1)}
				curFocus = 0
				app.SetFocus(focusables[curFocus])
			}
			currentPage = "tails"
			return nil
		case tcell.KeyF3:
			pages.SwitchToPage("proc")
			battleshell.UpdateFooter(footer, "Main (F1)", "Logs (F2)", "*Procs (F3)", "Conf (F4)")
			if flex, ok := procView.(*tview.Flex); ok && flex.GetItemCount() >= 2 {
				focusables = []tview.Primitive{procList, flex.GetItem(1)}
				curFocus = 0
				app.SetFocus(focusables[curFocus])
			}
			currentPage = "proc"
			return nil
		case tcell.KeyF4:
			pages.SwitchToPage("conf")
			battleshell.UpdateFooter(footer, "Main (F1)", "Logs (F2)", "Procs (F3)", "*Conf (F4)")
			focusables = []tview.Primitive{confPanel}
			curFocus = 0
			app.SetFocus(confPanel)
			currentPage = "conf"
			return nil
		}
		return ev
	})

	if root != nil {
		app.SetRoot(root, true)
	}
	if err := app.EnableMouse(false).Run(); err != nil {
		log.Fatal(err)
	}
}
