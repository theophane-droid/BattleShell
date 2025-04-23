package battleshell

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/rivo/tview"
)

/* ---------- CORRECTION : expression régulière ------------- */
/* simple échappement (« \{ »), pas double                    */
var argRegex = regexp.MustCompile(`\{([^}]+)\}`)

// BuildMenu construit récursivement un menu avec exécution de commandes
// (simples, à arguments nommés, ou avec SelectActions).
func BuildMenu(
	app *tview.Application,
	pages *tview.Pages,
	cfg MenuConfig,
	output *tview.TextArea,
	rootMenu *tview.Primitive,
	isRoot bool,
	inFormFlag *bool,
) tview.Primitive {

	/* ---------------- liste du menu ---------------- */
	list := tview.NewList()
	list.ShowSecondaryText(false).
		SetBorder(true).
		SetTitle(" " + cfg.Title + " ")

	if *rootMenu == nil && isRoot {
		*rootMenu = list
	}

	/* ------------ raccourcis 1..9 puis 0 ----------- */
	count := 0
	nextShortcut := func() rune {
		if count < 9 {
			r := rune('1' + count)
			count++
			return r
		}
		if count == 9 {
			count++
			return '0'
		}
		count++
		return 0
	}

	/* ---------------- helper run ------------------- */
	runCmd := func(command string, cmdCfg CommandConfig) {
		out, _ := exec.Command(bashPath, "-c", command).CombinedOutput()
		outText := string(out)

		/* ------ pas de SelectActions : texte brut ----- */
		if len(cmdCfg.SelectActions) == 0 {
			output.SetText(outText, true)
			return
		}

		/* ------ SelectActions : liste des lignes -------------- */
		lines := strings.Split(strings.TrimSpace(outText), "\n")

		prevPage := "main"
		pages.HidePage(prevPage)

		lineList := tview.NewList()
		lineList.ShowSecondaryText(false).
			SetBorder(true).
			SetTitle(" Select line ")

		for _, l := range lines {
			line := l // capture
			display := line
			if len(display) > 100 {
				display = display[:100]
			}
			lineList.AddItem(display, "", 0, func() {
				fields := strings.Fields(line)
				vals := map[string]string{}
				for i, f := range cmdCfg.Fields {
					if i < len(fields) {
						vals[f] = fields[i]
					}
				}

				/* ---------------- modal ------------------ */
				btns := make([]string, len(cmdCfg.SelectActions))
				for i, a := range cmdCfg.SelectActions {
					btns[i] = a.Name
				}
				btns = append(btns, "Back")

				modal := tview.NewModal().
					SetText("Choose action for:\n" + display).
					AddButtons(btns).
					SetDoneFunc(func(ix int, label string) {
						app.SetFocus(lineList)
						pages.RemovePage("actions_modal")

						if label == "Back" || ix >= len(cmdCfg.SelectActions) {
							return
						}
						act := cmdCfg.SelectActions[ix]
						final := act.Template
						for k, v := range vals {
							final = strings.ReplaceAll(final, "{"+k+"}", v)
						}
						output.SetText("", true)
						ExecuteCommand(final, output, app)
					})

				pages.AddPage("actions_modal", modal, true, true)
			})
		}

		lineList.AddItem("← Back to menu", "", 'b', func() {
			pages.RemovePage("line_selector")
			pages.ShowPage(prevPage)
			app.SetFocus(list)
		})

		pages.AddPage("line_selector", lineList, true, true)
		app.SetFocus(lineList)
	}

	/* ------------- commandes principales ------------ */
	for _, c := range cfg.Commands {
		cmd := c
		sc := nextShortcut()

		list.AddItem(cmd.Name, cmd.Description, sc, func() {
			tmpl := cmd.Command

			/* ---- arguments nommés ? ---- */
			if m := argRegex.FindAllStringSubmatch(tmpl, -1); len(m) > 0 {
				seen := map[string]bool{}
				var names []string
				for _, sub := range m {
					if !seen[sub[1]] {
						seen[sub[1]] = true
						names = append(names, sub[1])
					}
				}

				vals := map[string]string{}
				form := tview.NewForm()
				for _, n := range names {
					argName := n // capture
					form.AddInputField(argName, "", 20, nil,
						func(text string) { vals[argName] = text })
				}

				form.AddButton("Run", func() {
					final := tmpl
					for k, v := range vals {
						final = strings.ReplaceAll(final, "{"+k+"}", v)
					}
					app.SetFocus(list)
					pages.RemovePage("args")
					*inFormFlag = false
					runCmd(final, cmd)
				})

				form.AddButton("Cancel", func() {
					app.SetFocus(list)
					pages.RemovePage("args")
					*inFormFlag = false
				})

				form.SetBorder(true).SetTitle(" Arguments ")
				*inFormFlag = true
				pages.AddAndSwitchToPage("args", form, true)
				app.SetFocus(form)
				return
			}

			/* -------- exécution directe -------- */
			runCmd(tmpl, cmd)
		})
	}

	/* ---------------- sous-menus ------------------ */
	for _, sm := range cfg.Submenus {
		sub := sm
		sc := nextShortcut()
		list.AddItem(sub.Title, "", sc, func() {
			submenu := BuildMenu(app, pages, sub, output, rootMenu, false, inFormFlag)
			pages.AddAndSwitchToPage(sub.Title, submenu, true)
			app.SetFocus(submenu)
		})
	}

	/* ---------------- Setup ---------------------- */
	list.AddItem("⚙ Setup", "Customize shell path", 's', func() {
		*inFormFlag = true
		form := tview.NewForm().
			AddInputField("Bash path", bashPath, 40, nil,
				func(text string) { bashPath = text }).
			AddButton("Save", func() {
				app.SetFocus(list)
				pages.RemovePage("setup")
				*inFormFlag = false
			}).
			AddButton("Cancel", func() {
				app.SetFocus(list)
				pages.RemovePage("setup")
				*inFormFlag = false
			})
		form.SetBorder(true).SetTitle(" Setup ")
		pages.AddAndSwitchToPage("setup", form, true)
		app.SetFocus(form)
	})

	/* ------------- Quit / Back ------------------- */
	if isRoot {
		list.AddItem("❌ Exit", "Quit", 'q', func() { app.Stop() })
	} else {
		list.AddItem("🔙 Back", "Go back", 'b', func() {
			app.SetFocus(*rootMenu)
			pages.SwitchToPage("main")
		})
	}

	return list
}
