package battleshell

import (
	"context"
	"fmt"
	"os/exec"
	"io"
	"bufio"
	"time"
	"github.com/rivo/tview"
	"sync"
)

var bashPath = "/bin/bash"

// ---
var (
	muCurrent sync.Mutex
	curCmd    *exec.Cmd // commande en cours (nil si aucune)
)

/* ────────── exécution ────────── */

func ExecuteCommand(
	cmdLine string,
	out *tview.TextArea,
	app *tview.Application,
	async ...bool, // true par défaut
) {

	runAsync := true
	if len(async) == 1 {
		runAsync = async[0]
	}

	/* —— 1. Arrêter l’éventuelle commande active ——— */
	muCurrent.Lock()
	if curCmd != nil && curCmd.Process != nil {
		_ = curCmd.Process.Kill() // SIGKILL; on pourrait envoyer SIGINT sous Unix
	}
	muCurrent.Unlock()

	header := fmt.Sprintf("$ %s -c \"%s\"\n", bashPath, cmdLine)
	out.SetText(header, true)

	run := func() {
		cmd := exec.Command(bashPath, "-c", cmdLine)

		// mémoriser comme “courant”
		muCurrent.Lock()
		curCmd = cmd
		muCurrent.Unlock()

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		_ = cmd.Start()

		reader := bufio.NewReader(io.MultiReader(stdout, stderr))

		for {
			chunk, err := reader.ReadString('\n') // lit jusqu’au \n OU EOF
			if len(chunk) > 0 {
				app.QueueUpdateDraw(func() {
					prev := out.GetText() // false = ne récupère pas les tags
					out.SetText(prev+chunk, true)
				})
			}
			if err != nil { // EOF ou erreur
				if err != io.EOF {
					app.QueueUpdateDraw(func() {
						prev := out.GetText()
						out.SetText(prev+"\nError: "+err.Error()+"\n", true)
					})
				}
				break
			}
		}

		_ = cmd.Wait()

		/* nettoyer curCmd */
		muCurrent.Lock()
		if curCmd == cmd {
			curCmd = nil
		}
		muCurrent.Unlock()
	}

	if runAsync {
		go run()
	} else {
		run()
	}
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
