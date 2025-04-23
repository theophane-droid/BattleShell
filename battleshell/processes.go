package battleshell

import (
	"fmt"
	"github.com/rivo/tview"
)

func NewProcessPanel(app *tview.Application, cfg []ProcessConfig, events <-chan ProcEvent) (*tview.List, tview.Primitive) {
    list := tview.NewList()
	list.
        ShowSecondaryText(false).
        SetBorder(true).
        SetTitle(" Processes ")
    output := tview.NewTextView()
	output.
        SetBorder(true).
        SetTitle(" Process Output ")

    // stocker le dernier output de chaque process
    lastOutputs := make([][]byte, len(cfg))

    // initialisation des noms
    for _, pc := range cfg {
        list.AddItem(pc.Name, "", 0, nil)
    }

    // écoute en arrière-plan des événements
    go func() {
        for ev := range events {
            // mémoriser le dernier output
            lastOutputs[ev.Index] = ev.Output

            app.QueueUpdateDraw(func() {
                name := cfg[ev.Index].Name
                // mettre à jour la couleur dans la liste
                if ev.Error != nil {
                    list.SetItemText(ev.Index, fmt.Sprintf("[red]%s[::-]", name), "")
                } else {
                    list.SetItemText(ev.Index, fmt.Sprintf("[green]%s[::-]", name), "")
                }
                // si c'est l'item courant, afficher la sortie
                if list.GetCurrentItem() == ev.Index {
                    output.Clear()
                    output.Write(ev.Output)
                }
            })
        }
    }()

    // si l'utilisateur change de sélection, ré-afficher le dernier output
    list.SetChangedFunc(func(idx int, text, sec string, r rune) {
		output.Clear()
		output.Write(lastOutputs[idx])
	})

    flex := tview.NewFlex().
        AddItem(list, 0, 1, true).
        AddItem(output, 0, 2, false)

    return list, flex
}
