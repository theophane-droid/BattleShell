package battleshell

import (
	"fmt"
	"os"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// NewConfigEditor returns a panel with a multiline editable textarea
// + Ctrl-S save, Ctrl-R reload, Ctrl-Q quit callback.
func NewConfigEditor(app *tview.Application, path string, onReload func()) tview.Primitive {
	// TextArea (editable, multiline) – requires tview ≥0.6.0
	editor := tview.NewTextArea()
		editor.SetBorder(true).
		SetTitle(" Edit config.json ")

	status := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	// helpers ----------------------------------------------------------
	load := func() {
		data, err := os.ReadFile(path)
		if err != nil {
			status.SetText(fmt.Sprintf("[red]Failed to load: %v[-]", err))
			return
		}
		editor.SetText(string(data), true)
		status.SetText("[green]Config loaded (Ctrl+S to save)[-]")
	}

	save := func() {
		if err := os.WriteFile(path, []byte(editor.GetText()), 0644); err != nil {
			status.SetText(fmt.Sprintf("[red]Save error: %v[-]", err))
			return
		}
		status.SetText("[green]Config saved (Ctrl+R to reload into app)[-]")
	}

	// initial load
	load()

	// key bindings -----------------------------------------------------
	editor.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch {
		case ev.Key() == tcell.KeyCtrlS: // save
			save()
			return nil
		case ev.Key() == tcell.KeyCtrlR: // reload from disk + callback
			load()
			if onReload != nil {
				onReload()
			}
			return nil
		case ev.Key() == tcell.KeyCtrlQ: // quit editor → return to previous view
			app.SetFocus(nil) // let caller reset the focus
			return nil
		}
		return ev
	})

	// layout: editor (grow) + status (1 line)
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(editor, 0, 1, true).
		AddItem(status, 1, 0, false)

	return flex
}
