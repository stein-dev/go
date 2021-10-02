package libopenvpn

import (
	"os/exec"
	"fmt"
	"bufio"
	"strings"
	"syscall"

	"github.com/stein/simple-tunnel-openvpn-enc/liblog"
	
)

var (
	Loop = true
)

func Stop() {
	Loop = false
}

type Openvpn struct {
	ProxyHost string
	InjectPort string
	FileName string
	AuthFileName string
}

func (o *Openvpn) LogInfoSplit(message string, slice int, color string) {
	if Loop {
		liblog.LogInfoSplit(message, slice, "", color)
	}
}

func (o *Openvpn) LogInfo(message string, color string) {
	o.LogInfoSplit(message, 0, color)
}

func (o *Openvpn) Start() {
	command := exec.Command(
		"cmd", "/c", fmt.Sprintf(
			"extras\\openvpn --config %s --auth-user-pass %s " +
				"--route %s 255.255.255.255 net_gateway " +
				"--http-proxy 127.0.0.1 %s",
			o.FileName,
			o.AuthFileName,
			o.ProxyHost,
			o.InjectPort,
		),
	)

	stdout, err := command.StdoutPipe()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(stdout)
	go func() {
		var line string
		for Loop && scanner.Scan() {
			line = scanner.Text()

			if strings.Contains(line, "Initialization Sequence Completed") {
				o.LogInfo("Connected", liblog.Colors["G1"])

			} else if strings.Contains(line, "Connection reset") {	
				o.LogInfo("Reconnecting", liblog.Colors["P1"])

			} else if strings.Contains(line, "Attempting to establish TCP connection") {
				o.LogInfo("Connecting to the OpenVPN Server", liblog.Colors["Y1"])
				
			} else if strings.Contains(line, "TLS: Initial packet") {
				o.LogInfo("Initial Packet Received", liblog.Colors["Y1"])

			} else if strings.Contains(line, "Received control message") {
				o.LogInfo("Getting server configuration", liblog.Colors["Y1"])

			} else if strings.Contains(line, "TLS: Initial packet") {
				o.LogInfo("Initial Packet Received", liblog.Colors["Y1"])

			} else if strings.Contains(line, "Preserving previous TUN/TAP") {
				o.LogInfo("Preserving previous TUN/TAP", liblog.Colors["Y1"])

			} else if strings.Contains(line, "WFP engine opened") {
				o.LogInfo("Blocking DNS", liblog.Colors["Y1"])

			} else if strings.Contains(line, "TEST ROUTES") {
				o.LogInfo("Setting up routes", liblog.Colors["Y1"])

			} else if strings.Contains(line, "Access is denied") {
				o.LogInfo(
					"Fatal Error: Run application as administrator!",
					liblog.Colors["R1"],
				)
				break

			} //else {
			// 	o.LogInfoSplit(line[25:], 22, liblog.Colors["C1"])

			// }
		}

		command.Process.Signal(syscall.SIGTERM)
	}()

	command.Start()
	command.Wait()
}
