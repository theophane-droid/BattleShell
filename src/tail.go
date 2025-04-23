package main

import (
	"fmt"
	"time"
	"io/ioutil"
	"strings"
	"github.com/rivo/tview"
)

func readLastLines(path string, n int) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) <= n {
		return data, nil
	}
	last := lines[len(lines)-n:]
	return []byte(strings.Join(last, "\n")+"\n"), nil
}

func BuildTailView(app *tview.Application, files []string) tview.Primitive {
	tailList := tview.NewList()
	tailList.ShowSecondaryText(false).SetBorder(true).SetTitle(" Tails ")
	tailOutput := tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetWrap(true)
	tailOutput.SetBorder(true).SetTitle(" Tail Output ")

	var ticker *time.Ticker
	var quit chan struct{}

	for _, file := range files {
		f := file
		tailList.AddItem(f, "", 0, func() {
			if ticker != nil {
				ticker.Stop(); close(quit)
			}
			tailOutput.Clear()
			of, err := readLastLines(f, 20)
			if err != nil {
				tailOutput.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
			} else {
				tailOutput.Write(of)
			}
			quit = make(chan struct{})
			ticker = time.NewTicker(2 * time.Second)
			go func(fp string, t *time.Ticker, q chan struct{}) {
				for {
					select {
					case <-t.C:
						of, err := readLastLines(fp, 20)
						app.QueueUpdateDraw(func() {
							tailOutput.Clear()
							if err != nil {
								tailOutput.Write([]byte(fmt.Sprintf("Error: %v\n", err)))
							} else {
								tailOutput.Write(of)
							}
						})
					case <-q:
						return
					}
				}
			}(f, ticker, quit)
		})
	}

	return tview.NewFlex().AddItem(tailList, 0, 1, true).AddItem(tailOutput, 0, 2, false)
}
