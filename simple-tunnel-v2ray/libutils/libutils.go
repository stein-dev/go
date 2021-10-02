package libutils

import (
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

func ClearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func KillProcess(p *os.Process) {
	if p == nil {
		return
	}

	switch runtime.GOOS {
	case "windows":
		p.Kill()
	default:
		// p.Signal(syscall.SIGTERM)
		// p.Signal(os.Interrupt)
		p.Signal(os.Kill)
	}
}

type InterruptHandler struct {
	Done   chan bool
	Handle func()
}

func (i *InterruptHandler) Start() {
	i.Done = make(chan bool)

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-ch
		if i.Handle != nil {
			i.Handle()
		}
		i.Done <- true
	}()
}

func (i *InterruptHandler) Wait() {
	<-i.Done
	os.Exit(0)
}
