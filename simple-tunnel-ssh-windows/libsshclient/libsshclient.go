package libsshclient

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"


	"github.com/pigscanfly/liblog"
)

var (
	hostkey = ""
	Loop          = true
	DefaultConfig = &Config{
		Host:     "127.0.0.1",
		Port:     "22",
		Username: "pigscanfly",
		Password: "pigscanfly",
		Tunnel:   "plink",
	}
)

func Stop() {
	Loop = false
}

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	Tunnel   string
}

type SshClient struct {
	Config     *Config
	InjectPort string
	ListenPort string
	Verbose    bool
	Loop       bool
}

func (s *SshClient) LogInfo(message string, color string) {
	if Loop && s.Loop {
		liblog.LogInfo(message, "INFO", color)
	}
}

func (s *SshClient) Stop() {

	s.Loop = false


}

func (s *SshClient) Start() {
	s.LogInfo("Using Plink as Tunneling Tool", liblog.Colors["Y1"])
	s.LogInfo("Connecting to the remote server", liblog.Colors["C1"])

	for Loop && s.Loop {
		command := exec.Command(
			"cmd", "/c", fmt.Sprintf(
				"echo y | extras\\plink.exe -v -x -a -T -C -noagent -N -ssh %s@%s -P %s -pw %s "+
					"-D 1080",
				s.Config.Username,
				s.Config.Host,
				s.Config.Port,
				s.Config.Password,
			),
		)

		stderr, err := command.StderrPipe()
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(stderr)
		go func() {
			var line string
			for Loop && s.Loop && scanner.Scan() {
				line = scanner.Text()
				//s.LogInfo(line, liblog.Colors["G2"])

				if strings.Contains(line, "Local port 1080 SOCKS dynamic forwarding") {
					s.LogInfo("CONNECTED", liblog.Colors["G1"])

				} else if strings.Contains(line, "Address already in use") {
					s.LogInfo("Port used by another programs", liblog.Colors["R1"])
					s.Stop()

				} else {
					if s.Verbose {
						s.LogInfo(line, liblog.Colors["C1"])
					}
				}
			}

			command.Process.Signal(syscall.SIGTERM)
		}()

		command.Start()
		command.Wait()

		time.Sleep(200 * time.Millisecond)

		//command.Process.Kill()
		s.LogInfo("Reconnecting", liblog.Colors["Y1"])
	}
}

func (s *SshClient) StartBitvise() {
	s.LogInfo("Using Bitvise as Tunneling Tool", liblog.Colors["Y1"])
	s.LogInfo("Connecting to the remote server", liblog.Colors["C1"])
	command0 := exec.Command(
		"cmd", "/c", fmt.Sprintf(
			"echo s | extras\\stnlc.exe "+s.Config.Username+"@"+s.Config.Host+":"+s.Config.Port+" -pw="+s.Config.Password+
				" -proxyFwding=y -proxyListIntf=127.0.0.1 -proxyListPort=1080 -proxy=y -proxyType=http "+
				"-proxyServer=127.0.0.1 -proxyPort=8888 -ka -title=SimpleTunnel -pwKbdi=password",
		),
	)
	command0.Run()
	

	for Loop && s.Loop {
		command1 := exec.Command(
			"cmd", "/c", fmt.Sprintf(
				"extras\\stnlc.exe "+s.Config.Username+"@"+s.Config.Host+":"+s.Config.Port+" -pw="+s.Config.Password+
					" -proxyFwding=y -proxyListIntf=127.0.0.1 -proxyListPort=1080 -proxy=y -proxyType=http "+
					"-proxyServer=127.0.0.1 -proxyPort=8888 -ka -title=SimpleTunnel -pwKbdi=password -unat",
			),
		)

		stdout, err := command1.StdoutPipe()
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(stdout)
		go func() {
			var line string
			for Loop && s.Loop && scanner.Scan() {
				line = scanner.Text()
				//s.LogInfo(line, liblog.Colors["G2"])

				if strings.Contains(line, "Enabled SOCKS/HTTP") {
					s.LogInfo("CONNECTED", liblog.Colors["G1"])

				}  else {
					if s.Verbose {
						s.LogInfo(line, liblog.Colors["C1"])
					}
				}
			}

			command1.Process.Signal(syscall.SIGTERM)
			
		}()

		command1.Start()
		command1.Wait()

		

		time.Sleep(200 * time.Millisecond)
		//command1.Process.Kill()
		s.LogInfo("Reconnecting", liblog.Colors["Y1"])
	}
}

