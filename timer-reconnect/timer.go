package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func doReconnect(s string) {
	log.Println("Timer Reconnect Initiated")
	cmd := exec.Command("sh", "-c", "systemctl restart "+s)
	cmd.Run()
}

func startTimer(s string, t int) {
	for {
		time.Sleep(time.Duration(t) * time.Second)
		log.Println("Timer:", t)
		log.Println("Service:", s)
		go doReconnect(s)
	}
}

// func handler(w http.ResponseWriter, r *http.Request) {
//   fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
// }

func main() {

	fmt.Println("Reconnection Timer by @pigscanfly | Version: 0.0.2")
	servicePtr := flag.String("service", "", "Service to be restarted")
	timerPtr := flag.Int("timer", 0, "Reconnection Timer")
	flag.Parse()

	if *servicePtr == "" || *timerPtr == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	startTimer(*servicePtr, *timerPtr)

	// http.HandleFunc("/", handler)
	// http.ListenAndServe(":51080", nil)
}
