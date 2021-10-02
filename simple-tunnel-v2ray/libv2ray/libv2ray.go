package libv2ray

import (
	"os/exec"
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/stein/simple-tunnel-v2ray/libutils"
	"github.com/fatih/color"
)

var (
	Loop = true
)

func Stop() {
	Loop = false
}

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

func Log(message string, prefix string) {
	fmt.Printf("%s%s%s%s%s", "\r", "\033[K", message, Colors["CC"], prefix)
}

func LogColor(message string, color string) {
	messages := strings.Split(message, "\n")

	for _, value := range messages {
		Log(color+value, "\n")
	}
}

func LogInfo(message string, info string, color string) {
	datetime := time.Now()
	LogColor(
		fmt.Sprintf("[%.2d:%.2d:%.2d]%[5]s %[4]s::%[5]s %[6]s%[8]s",
			datetime.Hour(), datetime.Minute(), datetime.Second(),
			Colors["P1"], Colors["CC"], color,
			info, message),
		color,
	)
}

func Start() {

	for Loop {
		command := exec.Command(
			"cmd", "/c", fmt.Sprintf(
				"extras\\v2ray.exe -c config\\enc-config.json",
			),
		)

		stderr, err := command.StdoutPipe()
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(stderr)
		go func() {
			var line string
			for Loop && scanner.Scan() {

				line = scanner.Text()
				//LogInfo(line, "INFO", Colors["C1"])
				color.Set(color.FgGreen)
				fmt.Println(line)
			
			}

			libutils.KillProcess(command.Process)
		}()

		command.Start()
		command.Wait()
	}
}
