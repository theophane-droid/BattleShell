package battleshell

import (
	"context"
	"fmt"
	"os/exec"
	"time"
	"github.com/rivo/tview"
)

var bashPath = "/bin/bash"

// ------------------------------------------------------------------
// Helpers
// ------------------------------------------------------------------

func ExecuteCommand(cmd string, out *tview.TextArea) {
	header := fmt.Sprintf("$ %s -c \"%s\"\n", bashPath, cmd)

	res, err := exec.Command(bashPath, "-c", cmd).CombinedOutput()
	if err != nil {
		out.SetText(header+"Error:\n"+err.Error()+"\n"+string(res), true) // true = autoscroll
		return
	}

	out.SetText(header+string(res), true)
}
// ------------------------------------------------------------------
// Process watchers
// ------------------------------------------------------------------

type ProcEvent struct {
	Index  int    // process index
	Output []byte // last stdout/stderr
	Error  error  // nil if OK
}

func StartProcessWatchers(cfg []ProcessConfig, buf int) (events <-chan ProcEvent, cancel func()) {
	ch := make(chan ProcEvent, buf)
	ctx, stop := context.WithCancel(context.Background())

	for i, p := range cfg {
		go func(idx int, pc ProcessConfig) {
			t := time.NewTicker(time.Duration(pc.Interval) * time.Second)
			defer t.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					out, err := exec.Command(bashPath, "-c", pc.Command).CombinedOutput()
					select {
					case ch <- ProcEvent{Index: idx, Output: out, Error: err}:
					default:
						// si le buffer est plein on droppe l'event le plus ancien
						<-ch
						ch <- ProcEvent{Index: idx, Output: out, Error: err}
					}
				}
			}
		}(i, p)
	}

	return ch, stop
}

func IsFormChild(p tview.Primitive) bool {
	switch p.(type) {
	case *tview.InputField, *tview.Button, *tview.DropDown, *tview.Checkbox:
		return true
	default:
		return false
	}
}
