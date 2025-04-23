package battleshell

import (
	"strings"

	"github.com/rivo/tview"
)

// NewTabBar crée la barre d'onglets (sans contenu).
func NewTabBar() *tview.TextView {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	tv.SetBorder(true)
	return tv
}

// NewFooter crée le footer fixe (avec bordure).
func NewFooter() *tview.TextView {
    ft := tview.NewTextView().
        SetDynamicColors(true).
        SetTextAlign(tview.AlignLeft)       // ← aligner à gauche
    ft.SetBorder(true)
    ft.SetText("BATTLESHELL by *droid")
    return ft
}

func UpdateFooter(ft *tview.TextView, items ...string) {
    parts := make([]string, 0, len(items)/2)
    for i := 0; i < len(items); i += 1 {
        parts = append(parts, items[i])
    }
    ft.SetText("BATTLESHELL by *droid           " + strings.Join(parts, "  |  "))
}