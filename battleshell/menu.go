package battleshell

import (
	"regexp"
	"strings"

	"github.com/rivo/tview"
)

var argRegex = regexp.MustCompile(`\{([^}]+)\}`)

func BuildMenu(app *tview.Application, pages *tview.Pages,
	cfg MenuConfig, output *tview.TextArea,
	rootMenu *tview.Primitive, isRoot bool,
	inFormFlag *bool) tview.Primitive {
	list := tview.NewList()
	list.ShowSecondaryText(false).
		SetBorder(true).
		SetTitle(" " + cfg.Title + " ")

	count := 0
	nextShortcut := func() rune {
		if count < 9 {
			r := rune('1' + count)
			count++
			return r
		} else if count == 9 {
			count++
			return '0'
		}
		count++
		return 0
	}

	for _, cmd := range cfg.Commands {
		cmd := cmd
		sc := nextShortcut()
		list.AddItem(cmd.Name, cmd.Description, sc, func() {
			tmpl := cmd.Command
			matches := argRegex.FindAllStringSubmatch(tmpl, -1)
			if len(matches) > 0 {
				names := []string{}
				seen := map[string]bool{}
				for _, m := range matches {
					if !seen[m[1]] {
						seen[m[1]] = true
						names = append(names, m[1])
					}
				}
				values := map[string]string{}
				form := tview.NewForm()
				for _, n := range names {
					form.AddInputField(n, "", 20, nil, func(text string) { values[n] = text })
				}
				form.AddButton("Run", func() {
					final := tmpl
					for k, v := range values {
						final = strings.ReplaceAll(final, "{"+k+"}", v)
					}
					pages.RemovePage("args"); ExecuteCommand(final, output)
					*inFormFlag = false
					app.SetFocus(list)
				})
				form.AddButton("Cancel", func() { pages.RemovePage("args"); app.SetFocus(list); *inFormFlag = false })
				form.SetBorder(true).SetTitle(" Arguments ")
				*inFormFlag = true
				pages.AddAndSwitchToPage("args", form, true)
				app.SetFocus(form)
				return
			}
			ExecuteCommand(tmpl, output)
		})
	}

	for _, sub := range cfg.Submenus {
		sub := sub
		sc := nextShortcut()
		list.AddItem(sub.Title, "", sc, func() {
			submenu := BuildMenu(app, pages, sub, output, rootMenu, false, inFormFlag)
			pages.AddAndSwitchToPage(sub.Title, submenu, true)
			app.SetFocus(submenu)
		})
	}

	list.AddItem("⚙ Setup", "Customize shell path", 's', func() {
		*inFormFlag = true
		form := tview.NewForm().
			AddInputField("Bash path", bashPath, 40, nil, func(text string) { bashPath = text }).
			AddButton("Save", func() { pages.RemovePage("setup"); app.SetFocus(list) ; *inFormFlag = false;}).
			AddButton("Cancel", func() { pages.RemovePage("setup"); app.SetFocus(list) ; *inFormFlag = false; })
		form.SetBorder(true).SetTitle(" Setup ")
		pages.AddAndSwitchToPage("setup", form, true)
		app.SetFocus(form)
	})

	if isRoot {
		list.AddItem("❌ Exit", "Quit", 'q', func() { app.Stop() })
		*rootMenu = list
	} else {
		list.AddItem("🔙 Back", "Go back", 'b', func() { pages.SwitchToPage("main"); app.SetFocus(*rootMenu) })
	}
	return list
}
