package libsshclient

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"


	"github.com/stein/simple-tunnel-ssh-v1.1/m/liblog"
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
	Host     string
	Port     string
	Username string
	Password string
	Tunnel   string
}

func (s *SshClient) LogInfo(message string, color string) {
	if Loop && s.Loop {
		liblog.LogInfo(message, "INFO", color)
	}
}

func (s *SshClient) Stop() {

	s.Loop = false


}

func (s *SshClient) SetRegedit() {
	var regadd = "REG ADD HKCU\\SOFTWARE\\SimonTatham\\PuTTY\\Sessions\\"
	//var regadd2 = "REG ADD HKCU\\SOFTWARE\\9bis.com\\KiTTY\\Sessions\\"
	delSession := exec.Command(
		"cmd", "/c", fmt.Sprintf(
			"REG DELETE HKCU\\SOFTWARE\\SimonTatham\\PuTTY\\Sessions\\%s /f",
			s.Host,
		),
	)
	delSession.Run()
	addPresent := exec.Command("cmd", "/c", fmt.Sprintf(
		regadd + "%s /v Present /t REG_DWORD /d 1",
		s.Host,
		),
	)
	addPresent.Run()
	addHost := exec.Command("cmd", "/c", fmt.Sprintf(
		regadd + "%s /v HostName /t REG_SZ /d %s",
		s.Host,
		s.Host,
		),
	)
	addHost.Start()
	addPort := exec.Command("cmd", "/c", fmt.Sprintf(
		regadd + "%s /v PortNumber /t REG_DWORD /d %s",
		s.Host,
		s.Port,
		),
	)
	addPort.Start()
	addProxyHost := exec.Command("cmd", "/c", fmt.Sprintf(
		regadd + "%s /v ProxyHost /t REG_SZ /d 127.0.0.1",
		s.Host,
		),
	)
	addProxyHost.Start()
	addProxyPort := exec.Command("cmd", "/c", fmt.Sprintf(
		regadd + "%s /v ProxyPort /t REG_DWORD /d %s",
		s.Host,
		s.InjectPort,
		),
	)
	addProxyPort.Start()
	addProxyMethod := exec.Command("cmd", "/c", fmt.Sprintf(
		regadd + "%s /v ProxyMethod /t REG_DWORD /d 3",
		s.Host,
		),
	)
	addProxyMethod.Start()
	addProxyUsername := exec.Command("cmd", "/c", fmt.Sprintf(
		regadd + "%s /v ProxyUsername /t REG_SZ /d ",
		s.Host,
		),
	)
	addProxyUsername.Start()
	addProxyPassword := exec.Command("cmd", "/c", fmt.Sprintf(
		regadd + "%s /v ProxyPassword /t REG_DWORD /d ",
		s.Host,
		),
	)
	addProxyPassword.Start()
	addCipher := exec.Command("cmd", "/c", fmt.Sprintf(
		regadd + "%s /v Cipher /t REG_SZ /d blowfish",
		s.Host,
		),
	)
	addCipher.Start()
	
}

func (s *SshClient) Start() {
	s.LogInfo("Using Plink as Tunneling Tool", liblog.Colors["Y1"])
	s.LogInfo("Connecting to the remote server", liblog.Colors["C1"])

	for Loop && s.Loop {
		command := exec.Command(
			"cmd", "/c", fmt.Sprintf(
				"echo y | extras\\plink.exe -v -x -a -T -C -noagent -N -ssh %s@%s -P %s -pw %s "+
					"-D 1080",
				s.Username,
				s.Host,
				s.Port,
				s.Password,
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

				} else if strings.Contains(line, "Using username") {
					s.LogInfo("Sent Username", liblog.Colors["C1"])
					
				} else if strings.Contains(line, "Will use HTTP proxy") {
					//s.LogInfo("CONNECTED", liblog.Colors["G1"])
					continue
				} else if strings.Contains(line, "Leaving host lookup") {
					//s.LogInfo("CONNECTED", liblog.Colors["G1"])
					continue 
				} else if strings.Contains(line, "Looking up host") {
					//s.LogInfo("CONNECTED", liblog.Colors["G1"])
					continue 
				} else if strings.Contains(line, "Connecting to HTTP proxy") {
					//s.LogInfo("CONNECTED", liblog.Colors["G1"])
					continue 
				} else if strings.Contains(line, "Connecting to HTTP proxy") {
					//s.LogInfo("CONNECTED", liblog.Colors["G1"])
					continue 
				} else if strings.Contains(line, "Connecting to 127.0.0.1") {
					//s.LogInfo("CONNECTED", liblog.Colors["G1"])
					continue 
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

