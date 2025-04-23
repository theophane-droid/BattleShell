package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "os/exec"
    "regexp"
    "strings"
    "time"

    "github.com/gdamore/tcell/v2"
    "github.com/rivo/tview"
)

var (
    argRegex = regexp.MustCompile(`\{([^}]+)\}`)
    bashPath = "/bin/bash"
)

func executeCommand(cmdStr string, output *tview.TextView) {
    output.Clear()
    output.Write([]byte(fmt.Sprintf("$ %s -c \"%s\"\n", bashPath, cmdStr)))
    out, err := exec.Command(bashPath, "-c", cmdStr).CombinedOutput()
    if err != nil {
        output.Write([]byte(fmt.Sprintf("Error: %v\n%s\n", err, out)))
    } else {
        output.Write(out)
        output.Write([]byte("\n"))
    }
}

func readLastLines(path string, n int) ([]byte, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    content := string(data)
    lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
    if len(lines) <= n {
        return data, nil
    }
    last := lines[len(lines)-n:]
    return []byte(strings.Join(last, "\n") + "\n"), nil
}

func buildMenu(app *tview.Application, pages *tview.Pages, cfg MenuConfig, output *tview.TextView, rootMenu *tview.Primitive, isRoot bool) tview.Primitive {
    list := tview.NewList()
    list.ShowSecondaryText(false)
    list.SetBorder(true)
    list.SetTitle(" " + cfg.Title + " ")

    // Setup
    list.AddItem("⚙ Setup", "Customize shell path", 's', func() {
        form := tview.NewForm().
            AddInputField("Bash path", bashPath, 40, nil, func(text string) { bashPath = text }).
            AddButton("Save", func() {
                pages.RemovePage("setup")
                app.SetFocus(list)
            }).
            AddButton("Cancel", func() {
                pages.RemovePage("setup")
                app.SetFocus(list)
            })
        form.SetBorder(true)
        form.SetTitle(" Setup ")
        pages.AddAndSwitchToPage("setup", form, true)
        app.SetFocus(form)
    })

    // Commands
    for _, cmd := range cfg.Commands {
        cmd := cmd
        list.AddItem(cmd.Name, cmd.Description, 0, func() {
            tmpl := cmd.Command
            matches := argRegex.FindAllStringSubmatch(tmpl, -1)
            if len(matches) > 0 {
                names := []string{}
                seen := map[string]bool{}
                for _, m := range matches {
                    name := m[1]
                    if !seen[name] {
                        seen[name] = true
                        names = append(names, name)
                    }
                }
                values := map[string]string{}
                form := tview.NewForm()
                for _, n := range names {
                    form.AddInputField(n, "", 20, nil, func(text string) {
                        values[n] = text
                    })
                }
                form.AddButton("Run", func() {
                    final := tmpl
                    for n, v := range values {
                        final = strings.ReplaceAll(final, "{"+n+"}", v)
                    }
                    pages.RemovePage("args")
                    executeCommand(final, output)
                    app.SetFocus(list)
                })
                form.AddButton("Cancel", func() {
                    pages.RemovePage("args")
                    app.SetFocus(list)
                })
                form.SetBorder(true)
                form.SetTitle(" Arguments ")
                pages.AddAndSwitchToPage("args", form, true)
                app.SetFocus(form)
                return
            }
            executeCommand(tmpl, output)
        })
    }

    // Submenus
    for _, sub := range cfg.Submenus {
        sub := sub
        list.AddItem(sub.Title, "", 0, func() {
            submenu := buildMenu(app, pages, sub, output, rootMenu, false)
            pages.AddAndSwitchToPage(sub.Title, submenu, true)
            app.SetFocus(submenu)
        })
    }

    // Exit / Back
    if isRoot {
        list.AddItem("❌ Exit", "Quit", 'q', func() { app.Stop() })
        *rootMenu = list
    } else {
        list.AddItem("🔙 Back", "Go back", 'b', func() {
            pages.SwitchToPage("main")
            app.SetFocus(*rootMenu)
        })
    }
    return list
}

func buildTailView(app *tview.Application, files []string, tailTicker **time.Ticker, tailQuit *chan struct{}) tview.Primitive {
    tailList := tview.NewList()
    tailList.ShowSecondaryText(false)
    tailList.SetBorder(true)
    tailList.SetTitle(" Tails ")
    tailOutput := tview.NewTextView().
        SetDynamicColors(true).
        SetScrollable(true).
        SetWrap(true)
    tailOutput.SetBorder(true)
    tailOutput.SetTitle(" Tail Output ")

    for _, f := range files {
        file := f
        tailList.AddItem(file, "", 0, func() {
            if *tailTicker != nil {
                (*tailTicker).Stop()
                close(*tailQuit)
                *tailTicker = nil
            }
            tailOutput.Clear()
            out, err := readLastLines(file, 20)
            if err != nil {
                tailOutput.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
            }
            tailOutput.Write(out)
            quit := make(chan struct{})
            *tailQuit = quit
            ticker := time.NewTicker(2 * time.Second)
            *tailTicker = ticker
            go func(f string, t *time.Ticker, q chan struct{}) {
                for {
                    select {
                    case <-t.C:
                        out, err := readLastLines(f, 20)
                        app.QueueUpdateDraw(func() {
                            tailOutput.Clear()
                            if err != nil {
                                tailOutput.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
                            }
                            tailOutput.Write(out)
                        })
                    case <-q:
                        return
                    }
                }
            }(file, ticker, quit)
        })
    }

    flex := tview.NewFlex().
        AddItem(tailList, 0, 1, true).
        AddItem(tailOutput, 0, 2, false)
    return flex
}

func main() {
    cfg, err := LoadConfig("config.json")
    if err != nil {
        log.Fatalf("Cannot load config: %v", err)
    }

    app := tview.NewApplication()

    tabs := tview.NewTextView().
        SetDynamicColors(true).
        SetText("[::b][F1] Main[::-]  [::b][F2] Tails[::-]").
        SetTextAlign(tview.AlignCenter)
    tabs.SetBorder(true)

    menuPages := tview.NewPages()
    output := tview.NewTextView().
        SetDynamicColors(true).
        SetScrollable(true).
        SetWrap(true)
    output.SetBorder(true).SetTitle(" Output ")
    var rootMenu tview.Primitive
    mainMenu := buildMenu(app, menuPages, cfg.Menu, output, &rootMenu, true)
    menuPages.AddAndSwitchToPage("main", mainMenu, true)
    input := tview.NewInputField()
    input.
        SetLabel("> ").
        SetFieldWidth(0).
        SetDoneFunc(func(key tcell.Key) {
            if key == tcell.KeyEnter {
                cmd := input.GetText()
                input.SetText("")
                executeCommand(cmd, output)
                app.SetFocus(rootMenu)
            }
        })
    input.SetBorder(true).SetTitle(" Shell Input ")

    mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
        AddItem(tview.NewFlex().
            AddItem(menuPages, 0, 1, true).
            AddItem(output, 0, 2, false), 0, 1, true).
        AddItem(input, 3, 0, false)

    var tailTicker *time.Ticker
    var tailQuit chan struct{}
    tailView := buildTailView(app, cfg.TailFiles, &tailTicker, &tailQuit)

    rootPages := tview.NewPages().
        AddPage("mainView", mainFlex, true, true).
        AddPage("tailView", tailView, true, false)

    // Footer text
    footer := tview.NewTextView().
        SetTextAlign(tview.AlignCenter).
        SetText("BATTLE SHELL by *droid")

    root := tview.NewFlex().SetDirection(tview.FlexRow).
        AddItem(tabs, 1, 0, false).
        AddItem(rootPages, 0, 1, true).
        AddItem(footer, 1, 0, false)

    app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        switch event.Key() {
        case tcell.KeyF1:
            if tailTicker != nil {
                tailTicker.Stop()
                close(tailQuit)
                tailTicker = nil
            }
            rootPages.SwitchToPage("mainView")
            tabs.SetText("[::b][F1] Main[::-]  [F2] Tails")
            app.SetFocus(rootMenu)
            return nil
        case tcell.KeyF2:
            rootPages.SwitchToPage("tailView")
            tabs.SetText("[F1] Main  [::b][F2] Tails[::-]")
            return nil
        }
        return event
    })

    if err := app.SetRoot(root, true).EnableMouse(true).Run(); err != nil {
        log.Fatal(err)
    }
}
