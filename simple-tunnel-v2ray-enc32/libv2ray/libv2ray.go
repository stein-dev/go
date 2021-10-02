package libv2ray

import (
	"os/exec"
	"bufio"
	"fmt"
	"strings"
	"log"

	"github.com/stein/simple-tunnel-v2ray-enc/libutils"
)

var (
	Loop = true	
	ConfigName1 = ""
)

type V2RayClient struct {
	ConfigName string
}

func Stop() {
	Loop = false
}

func (v *V2RayClient) Start() {

	if ( v.ConfigName == "") {
		log.Println("Error loading config.Restart app.")
		return
	}

	log.Println("Loading Config...")
	log.Println("Starting V2Ray...")

	for Loop {
		
		command := exec.Command(
			"cmd", "/c", fmt.Sprintf(
				"extras\\v2ray.exe -c %s",
				 v.ConfigName,
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
				
				if strings.Contains(line, "Reading config") {
					continue
				} else if strings.Contains(line, "V2Fly") {
					continue
				} else if strings.Contains(line, "A unified platform") {
					continue
				} else if strings.Contains(line, "[Warning] v2ray.com/core") {
					log.Println("V2Ray Started")
		   		} else if strings.Contains(line, "failed to dial") {
					log.Println("Failed to connect to the server")
		   		} else if strings.Contains(line, "\\AppData\\Local\\Temp") {
					break
		   		} else if strings.Contains(line, "AppData") {
					break
		   		} else if strings.Contains(line, "ApplicationData") {
					break
		   		} else if strings.Contains(line, "Application Data") {
					break
		   		} else if strings.Contains(line, "failed to listen TCP") {
					log.Println("Listen port is being used by another program.")
					break
		   		} else {
					fmt.Println(line)
				}	
				
				
				// else if strings.Contains(line, "accepted") {
				// 	g.Println(line)
				// }   else if strings.Contains(line, "failed to process outbound traffic") {
				// 	r.Println(line)
				// } else if strings.Contains(line, "[Info]") {
				// 	c.Println(line)
				// } else if strings.Contains(line, "[Warning]") {
				// 	r.Println(line)
				// }  else {
				// 	fmt.Println(line)
				// }
			}
			
			libutils.KillProcess(command.Process, v.ConfigName)
			
		}()

		command.Start()
		command.Wait()
	}
}
