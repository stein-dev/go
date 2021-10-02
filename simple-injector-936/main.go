package main

import (
	"io"
	"net"
	"strings"
	"time"
	"fmt"
	"os/exec"
	"os"

	"github.com/stein/simple-injector/libinject"
	"github.com/stein/simple-injector/libutils"

)

type ClientManager struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	payload    string
	auth       string
}

type Client struct {
	socket net.Conn
	proxy  Proxy
	timer  *time.Timer
}

type Proxy struct {
	socket    net.Conn
	connected bool
}

type Config struct {
	InjectorSettings *libinject.Config

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


func (manager *ClientManager) start() {
	for {
		select {
		case connection := <-manager.register:
			manager.clients[connection] = true
			//log.Println("Client Connected:", connection.socket.RemoteAddr())
			LogInfo("Client Connected", "", Colors["B1"])
		case connection := <-manager.unregister:
			delete(manager.clients, connection)
			//log.Println("Client Disconnected:", connection.socket.RemoteAddr())
			LogInfo("Client Disconnected", "", Colors["R1"])
		}
	}
}

func (manager *ClientManager) parsePayload(request *[]byte) []byte {
	if isHttpRequest(request) {
		reqString := string(*request)
		payload := manager.payload                          // copy payload from manager
		splitRequestRaw := strings.Split(reqString, "\r\n") // split http request
		splitRequest := strings.Split(splitRequestRaw[0], " ")

		connHost := splitRequest[1]  // ip after CONNECT
		connProto := splitRequest[2] // http protocol

		parsed := strings.ReplaceAll(payload, "[host_port]", connHost)

		if manager.auth == "" { // authentication
			parsed = strings.ReplaceAll(parsed, "[protocol]", connProto)
		}

		parsed = strings.ReplaceAll(parsed, "[ua]", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.111 Safari/537.36")
		parsed = strings.ReplaceAll(parsed, "[crlf]", "\r\n")
		parsed = strings.ReplaceAll(parsed, "[cr]", "\r")
		parsed = strings.ReplaceAll(parsed, "[lf]", "\n")

		return []byte(parsed)
	}
	return *request
}

func isHttpRequest(request *[]byte) bool {
	if strings.Contains(string(*request), "CONNECT") ||
		strings.Contains(string(*request), "GET") ||
		strings.Contains(string(*request), "POST") ||
		strings.Contains(string(*request), "PUT") ||
		strings.Contains(string(*request), "OPTIONS") ||
		strings.Contains(string(*request), "TRACE") ||
		strings.Contains(string(*request), "TRACE") ||
		strings.Contains(string(*request), "OPTIONS") ||
		strings.Contains(string(*request), "TRACE") ||
		strings.Contains(string(*request), "PATCH") ||
		strings.Contains(string(*request), "DELETE") {
		return true
	}
	return false
}

func (manager *ClientManager) handleConnection(client *Client) {

	// close connection after timer
	go func() {
		if client.timer != nil {
			<-client.timer.C
			client.proxy.socket.Close()
			client.socket.Close()
		}
	}()

	// client to proxy
	go func() {
		size := 32 * 1024
		var r io.Reader = client.socket

		if l, ok := r.(*io.LimitedReader); ok && int64(size) > l.N {
			if l.N < 1 {
				size = 1
			} else {
				size = int(l.N)
			}
		}
		buf := make([]byte, size)

		for {
			nr, err := r.Read(buf)
			

			if nr > 0 {

				if client.proxy.connected == false {
					p := manager.parsePayload(&buf) // parse connection
					_, err = client.proxy.socket.Write(p)
					if err != nil {
						// cant write to proxy
						//log.Println(err)
						LogInfo("Unexpected error [1]", "", Colors["R1"])
						break
					}
					client.proxy.connected = true
					

				} else {
					nw, err := client.proxy.socket.Write(buf[0:nr])
					if err != nil {
						//log.Println(err)
						LogInfo("Unexpected error [2]", "", Colors["R1"])
						break
					}

					if nr != nw {
						//log.Println(err)
						LogInfo("Unexpected error [3]", "", Colors["R1"])
						break
					}
				}
			}

			if err != nil {
				if err != io.EOF {
					//log.Println(err)
					//log.Println("Server unexpectedly closed network connection")
					LogInfo("Server unexpectedly closed network connection", "", Colors["R1"])
				}
				manager.unregister <- client
				_ = client.proxy.socket.Close()
				break
			}
		}
	}()

	// proxy to client
	go func() {
		size := 32 * 1024
		var r io.Reader = client.proxy.socket

		if l, ok := r.(*io.LimitedReader); ok && int64(size) > l.N {
			if l.N < 1 {
				size = 1
			} else {
				size = int(l.N)
			}
		}
		buf := make([]byte, size)

		LogInfo("Injecting Payload", "", Colors["P1"])
		for {
			nr, err := r.Read(buf)
			data := ""
			data += string(buf[:nr])
			
			if nr > 0 {
				nw, err := client.socket.Write(buf[0:nr])

				if strings.Contains(strings.Split(data, "\r\n")[0], "HTTP/1.1 200 Connection established") {
					//log.Println("HTTP/1.1 200 Connection established")
					LogInfo("HTTP/1.1 200 Connection established", "", Colors["B1"])
				}
				
				if err != nil {
					break
				}
				if nr != nw {
					break
				}
			}

			if err != nil {
				_ = client.socket.Close()
				break
			}
		}
	}()
}

func ClearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {

	ClearScreen()
	LogInfo("Simple Injector 936 | @pigscanfly", "", Colors["B1"])
	
	config := new(Config)
	defaultConfig := new(Config)
	defaultConfig.InjectorSettings = libinject.DefaultConfig

	libutils.JsonReadWrite(libutils.RealPath("inject.json"), config, defaultConfig)

	Inject := new(libinject.Inject)
	Inject.Config = config.InjectorSettings

	hostPtr := Inject.Config.ProxyHost
	portPtr := Inject.Config.ProxyPort
	payloadPtr := Inject.Config.Payload
	listenPtr := Inject.Config.ListenPort
	timerPtr := Inject.Config.Timer

	LogInfo("Injector running on port: " + listenPtr, "", Colors["B1"])
	LogInfo("Remote proxy running on: " + hostPtr + ":" + portPtr, "", Colors["B1"])
	LogInfo("Payload: " + payloadPtr, "", Colors["B1"])

	conn, err := net.Listen("tcp", "127.0.0.1:" + listenPtr)
	LogInfo("Waiting for client to connect", "", Colors["B1"])

	if err != nil {
		LogInfo("Unexpected error [4]", "", Colors["R1"])	
	}

	manager := ClientManager{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		payload:    string(payloadPtr),
	}

	go manager.start()


	for {

		accept, err := conn.Accept()
		if err != nil {
			LogInfo("Unexpected error. Restart application", "", Colors["R1"])
		}

		proxy, err := net.Dial("tcp", hostPtr + ":" + portPtr)
		if err != nil {
			LogInfo("Unable to connect to the remote proxy server", "", Colors["R1"])
			continue
		}

		client := &Client{
			socket: accept,
			proxy: Proxy{
				socket:    proxy,
				connected: false,
			},
			timer: nil,
		}

		if timerPtr != 0 {
			client.timer = time.NewTimer(time.Duration(timerPtr) * time.Second)
			LogInfo("Timer Reconnect Initiated", "", Colors["P1"])
		}

		

		manager.register <- client
		go manager.handleConnection(client)

	}

}
