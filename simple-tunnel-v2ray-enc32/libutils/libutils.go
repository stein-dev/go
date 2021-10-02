package libutils

import (
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"time"

)

func ClearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func KillProcess(p *os.Process, a string) {
	if p == nil {
		return
	}

	switch runtime.GOOS {
	case "windows":
		p.Kill()
		time.Sleep(2)
		os.RemoveAll(a)

	default:
		// p.Signal(syscall.SIGTERM)
		// p.Signal(os.Interrupt)
		p.Signal(os.Kill)
		time.Sleep(2)
		os.RemoveAll(a)
	}
}

type InterruptHandler struct {
	Done   chan bool
	Handle func()
}

func (i *InterruptHandler) Start() {
	i.Done = make(chan bool)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	go func() {
		<-ch
		if i.Handle != nil {
			i.Handle()		
		}
		
		i.Done <- true
	}()
}

func (i *InterruptHandler) Wait(a string) {
	<-i.Done
	os.RemoveAll(a)
	time.Sleep(2)
	os.Exit(1)
}
