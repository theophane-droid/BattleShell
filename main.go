package main

import (
	"log"

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
		rootMenu tview.Primitive

		menuPages *tview.Pages
		tailView  tview.Primitive
		procView  tview.Primitive
		procList tview.Primitive
		confPanel tview.Primitive
		mainView  tview.Primitive
		pages     *tview.Pages
		root      *tview.Flex

		focusables []tview.Primitive
		curFocus   int

		procEvents     <-chan battleshell.ProcEvent
		cancelWatchers func()

		reloadEverything func()
	)

	/* ────────── persistent widgets ────── */
	output := tview.NewTextArea()
	output.SetText("", true)
	output.SetBorder(true).SetTitle(" Output ")

	input := tview.NewInputField()
	input.SetLabel("> ").SetFieldWidth(0).
		SetBorder(true).SetTitle(" Shell Input ")
	input.SetDoneFunc(func(k tcell.Key) {
		if k == tcell.KeyEnter {
			cmd := input.GetText()
			input.SetText("")
			battleshell.ExecuteCommand(cmd, output)
		}
	})

	/* ────────── reloadEverything ──────── */
	reloadEverything = func() {
		if cancelWatchers != nil {
			cancelWatchers()
		}

		cfg, err := battleshell.LoadConfig("config.json")
		if err != nil {
			app.SetRoot(
				tview.NewModal().
					SetText("Invalid config.json:\n\n"+err.Error()).
					AddButtons([]string{"OK"}),
				true)
			return
		}

		procEvents, cancelWatchers = battleshell.StartProcessWatchers(cfg.Processes, len(cfg.Processes)*2)

		/* rebuild MAIN page */
		rootMenu = nil
		menuPages = tview.NewPages()
		mainMenu  := battleshell.BuildMenu(app, menuPages, cfg.Menu, output, &rootMenu, true, &inForm)
		menuPages.AddPage("main", mainMenu, true, true)

		mainView = tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().
				AddItem(menuPages, 0, 1, true).
				AddItem(output,     0, 2, false), 0, 1, true).
			AddItem(input, 3, 0, false)

		/* rebuild other pages */
		tailView           = battleshell.BuildTailView(app, cfg.TailFiles)
		procList, procView = battleshell.NewProcessPanel(app, cfg.Processes, procEvents)
		confPanel          = battleshell.NewConfigEditor(app, "config.json", reloadEverything)

		pages = tview.NewPages().
			AddPage("main",  mainView, true,  true).
			AddPage("tails", tailView, true,  false).
			AddPage("proc",  procView, true,  false).
			AddPage("conf",  confPanel, true, false)

		if root == nil {
			root = tview.NewFlex().SetDirection(tview.FlexRow)
		}
		root.Clear().
			AddItem(tabBar, 1, 0, false).
			AddItem(pages, 0, 1, true).
			AddItem(footer, 3, 0, false)

		focusables = []tview.Primitive{menuPages, output, input}
		curFocus   = 0

		app.SetRoot(root, true)
		app.SetFocus(focusables[curFocus])
		battleshell.UpdateFooter(footer, "*Main (F1)", "Logs (F2)", "Procs (F3)", "Conf (F4)")
	}

	/* first load */
	reloadEverything()

	/* ────────── Global keys ───────────── */
	app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if inForm {          // laisser TAB interne au form
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
			return nil
		case tcell.KeyF2:
			pages.SwitchToPage("tails")
			battleshell.UpdateFooter(footer, "Main (F1)", "*Logs (F2)", "Procs (F3)", "Conf (F4)")
			if flex, ok := tailView.(*tview.Flex); ok && flex.GetItemCount() >= 2 {
				focusables = []tview.Primitive{flex.GetItem(0), flex.GetItem(1)}
				curFocus = 0
				app.SetFocus(focusables[curFocus])
			}
			return nil
		case tcell.KeyF3:
			pages.SwitchToPage("proc")
			battleshell.UpdateFooter(footer, "Main (F1)", "Logs (F2)", "*Procs (F3)", "Conf (F4)")
			if flex, ok := procView.(*tview.Flex); ok && flex.GetItemCount() >= 2 {
				focusables = []tview.Primitive{procList, flex.GetItem(1)}
				curFocus = 0
				app.SetFocus(focusables[curFocus])
			}
			return nil
		case tcell.KeyF4:
			pages.SwitchToPage("conf")
			battleshell.UpdateFooter(footer, "Main (F1)", "Logs (F2)", "Procs (F3)", "*Conf (F4)")
			focusables = []tview.Primitive{confPanel}
			curFocus = 0
			app.SetFocus(confPanel)
			return nil
		/* …F1 / F2 / F3 / F4 inchangés… */
		}
		return ev
	})
	

	if root != nil {
		app.SetRoot(root, true)
	}
	if err := app.EnableMouse(false).Run(); err != nil {
		log.Fatal(err)
	}
	
	/* run */
	// if err := app.SetRoot(root, true).EnableMouse(false).Run(); err != nil {
	// 	log.Fatal(err)
	// }
}