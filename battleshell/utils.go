package battleshell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/rivo/tview"
)

/* ----------------------------------------------------------- */
/*                   Streaming  +  Process kill                */
/* ----------------------------------------------------------- */

var (
	bashPath = "/bin/bash"

	muCurrent sync.Mutex
	curCmd    *exec.Cmd // commande en cours (nil si aucune)
)

/* commandes qui exigent un TTY (nc, ssh, ftp, mysql, …) */
func interactive(cmd string) bool {
	kw := []string{"nc ", "ssh ", "ftp ", "mysql ", "psql "}
	for _, k := range kw {
		if strings.Contains(cmd, k) {
			return true
		}
	}
	return false
}

/* Exécute une commande ; si interactive ⇒ pty, sinon pipes. */
func ExecuteCommand(
	cmdLine string,
	out *tview.TextArea,
	app *tview.Application,
	async ...bool,
) {
	runAsync := true
	if len(async) == 1 {
		runAsync = async[0]
	}

	/* 1. stop previous */
	muCurrent.Lock()
	if curCmd != nil && curCmd.Process != nil {
		_ = curCmd.Process.Kill()
	}
	muCurrent.Unlock()

	header := fmt.Sprintf("$ %s -c \"%s\"\n", bashPath, cmdLine)
	out.SetText(header, true)

	run := func() {
		cmd := exec.Command(bashPath, "-c", cmdLine)

		muCurrent.Lock()
		curCmd = cmd
		muCurrent.Unlock()
		fmt.Println(cmdLine)
		if true {
			ptmx, _ := pty.Start(cmd)
			buf := make([]byte, 4096)

			for {
				n, err := ptmx.Read(buf)
				if n > 0 {
					chunk := string(buf[:n])
					app.QueueUpdateDraw(func() {
						prev:= out.GetText()
						out.SetText(prev+chunk, true)
					})
				}
				if err != nil {
					break
				}
			}
			_ = cmd.Wait()
			_ = ptmx.Close()

		} else {
			/* ----------- non-interactive ---------- */
			stdout, _ := cmd.StdoutPipe()
			stderr, _ := cmd.StderrPipe()
			_ = cmd.Start()

			reader := bufio.NewReader(io.MultiReader(stdout, stderr))
			buf := make([]byte, 4096)

			for {
				n, err := reader.Read(buf)
				if n > 0 {
					chunk := string(buf[:n])
					app.QueueUpdateDraw(func() {
						prev:= out.GetText()
						out.SetText(prev+chunk, true)
					})
				}
				if err != nil {
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
		}

		/* cleanup */
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

/* ----------------------------------------------------------- */
/*                    Process watchers                         */
/* ----------------------------------------------------------- */

type ProcEvent struct {
	Index  int
	Output []byte
	Error  error
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
						<-ch // drop oldest
						ch <- ProcEvent{Index: idx, Output: out, Error: err}
					}
				}
			}
		}(i, p)
	}

	return ch, stop
}

/* ----------------------------------------------------------- */
/*                  Helpers Tview                              */
/* ----------------------------------------------------------- */

func IsFormChild(p tview.Primitive) bool {
	switch p.(type) {
	case *tview.InputField, *tview.Button, *tview.DropDown, *tview.Checkbox:
		return true
	default:
		return false
	}
}
