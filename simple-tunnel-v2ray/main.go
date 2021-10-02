package main

import (
	"syscall"
	"unsafe"
	"fmt"
	//"strings"
	//"time"
	//"io/ioutil"
	//"os"
	//"log"

	"github.com/stein/simple-tunnel-v2ray/libutils"
	"github.com/stein/simple-tunnel-v2ray/libv2ray"
)

var (
	InterruptHandler = new(libutils.InterruptHandler)
)

var (
	Colors = map[string]string{
		"R1": "\033[31;1m", "R2": "\033[31;2m",
		"G1": "\033[32;1m", "G2": "\033[32;2m",
		"Y1": "\033[33;1m", "Y2": "\033[33;2m",
		"B1": "\033[34;1m", "B2": "\033[34;2m",
		"P1": "\033[35;1m", "P2": "\033[35;2m",
		"C1": "\033[36;1m", "C2": "\033[36;2m", "CC": "\033[0m",
	}
)


func init() {
	InterruptHandler.Handle = func() {
		libv2ray.Stop()
		//liblog.LogKeyboardInterrupt()
	}
	InterruptHandler.Start()
}

func SetConsoleTitle(title string) (int, error) {
	handle, err := syscall.LoadLibrary("Kernel32.dll")
	if err != nil {
		return 0, err
	}
	defer syscall.FreeLibrary(handle)
	proc, err := syscall.GetProcAddress(handle, "SetConsoleTitleW")
	if err != nil {
		return 0, err
	}
	r, _, err := syscall.Syscall(proc, 1, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))), 0, 0)
	return int(r), err
}

func main() {
	header := `
	
	░██████╗██╗███╗░░░███╗██████╗░██╗░░░░░███████╗  ████████╗██╗░░░██╗███╗░░██╗███╗░░██╗███████╗██╗░░░░░
	██╔════╝██║████╗░████║██╔══██╗██║░░░░░██╔════╝  ╚══██╔══╝██║░░░██║████╗░██║████╗░██║██╔════╝██║░░░░░
	╚█████╗░██║██╔████╔██║██████╔╝██║░░░░░█████╗░░  ░░░██║░░░██║░░░██║██╔██╗██║██╔██╗██║█████╗░░██║░░░░░
	░╚═══██╗██║██║╚██╔╝██║██╔═══╝░██║░░░░░██╔══╝░░  ░░░██║░░░██║░░░██║██║╚████║██║╚████║██╔══╝░░██║░░░░░
	██████╔╝██║██║░╚═╝░██║██║░░░░░███████╗███████╗  ░░░██║░░░╚██████╔╝██║░╚███║██║░╚███║███████╗███████╗
	╚═════╝░╚═╝╚═╝░░░░░╚═╝╚═╝░░░░░╚══════╝╚══════╝  ░░░╚═╝░░░░╚═════╝░╚═╝░░╚══╝╚═╝░░╚══╝╚══════╝╚══════╝
	`
	author := `▀▄▀▄▀▄ Version 0.1.0 (c) 2021 @pigscanfly ▄▀▄▀▄▀ `
		

	SetConsoleTitle("Simple Tunnel V2RAY | @pigscanfly")
	fmt.Println(header)
	fmt.Println("\t\t\t" + author)
	fmt.Println(" ")

	libv2ray.Start()

	InterruptHandler.Wait()
}
