package main

import (
	"io"
	"net"
	"strings"
	"time"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"bufio"
	"encoding/json"
	"syscall"
	"unsafe"
	"os/exec"
	"strconv"

	"github.com/stein/simple-tunnel-ssh/libinject"
	"github.com/stein/simple-tunnel-ssh/libutils"
	"github.com/stein/simple-tunnel-ssh/libsshclient"
	"github.com/stein/simple-tunnel-ssh/liblog"
	//"github.com/common-nighthawk/go-figure"

)

var isLock bool

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
	SshSettings *libsshclient.Config

}

func (manager *ClientManager) start() {
	for {
		select {
		case connection := <-manager.register:
			manager.clients[connection] = true
			//log.Println("Client Connected:", connection.socket.RemoteAddr())
			liblog.LogInfo("Client Connected", "", liblog.Colors["G1"])
		case connection := <-manager.unregister:
			delete(manager.clients, connection)
			//log.Println("Client Disconnected:", connection.socket.RemoteAddr())
			liblog.LogInfo("Client Disconnected", "", liblog.Colors["R1"])
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
			// client.proxy.socket.Close()
			// client.socket.Close()
			cmd0 := exec.Command("cmd", "/c", "taskkill /f /im plink.exe")
			cmd0.Start()
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
						liblog.LogInfo("Unexpected error [1]", "", liblog.Colors["R1"])
						break
					}
					client.proxy.connected = true
					

				} else {
					nw, err := client.proxy.socket.Write(buf[0:nr])
					if err != nil {
						//log.Println(err)
						liblog.LogInfo("Unexpected error [2]", "", liblog.Colors["R1"])
						break
					}

					if nr != nw {
						//log.Println(err)
						liblog.LogInfo("Unexpected error [3]", "", liblog.Colors["R1"])
						break
					}
				}
			}

			if err != nil {
				if err != io.EOF {
					//log.Println(err)
					//log.Println("Server unexpectedly closed network connection")
					liblog.LogInfo("Server unexpectedly closed network connection", "", liblog.Colors["R1"])
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
		//d := json.NewDecoder(r)
		//d := json.NewDecoder(client.proxy.socket)
		liblog.LogInfo("Injecting Payload", "", liblog.Colors["P1"])
		for {
			nr, err := r.Read(buf)
			data := ""
			data += string(buf[:nr])
			
			if nr > 0 {
				nw, err := client.socket.Write(buf[0:nr])

				if strings.Contains(strings.Split(data, "\r\n")[0], "HTTP/1.1 200") {
					liblog.LogInfo("HTTP/1.1 200 Connection established", "", liblog.Colors["G1"])
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

func readline() string {
	bio := bufio.NewReader(os.Stdin)
	line, _, err := bio.ReadLine()
	if err != nil {
		fmt.Println(err)
	}
	return string(line)
}

func writeToFile(data, file string) {
	ioutil.WriteFile(file, []byte(data), 777)
}

func createFile(file string) {
	os.Create(file)
}

func readFromFile(file string) ([]byte, error) {
	data, err := ioutil.ReadFile(file)
	return data, err
}

func decrypt(encryptedString string, keyString string) (decryptedString string) {

	key, _ := hex.DecodeString(keyString)
	enc, _ := hex.DecodeString(encryptedString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return fmt.Sprintf("%s", plaintext)
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

type ConfigFile struct {
	InjectorSettings InjectorKeys
	SshSettings SshKeys
}

type InjectorKeys struct {
	ListenPort	  string
	Payload	string
	ProxyHost     string
	ProxyPort     string
	Username string
	Password string
	Timer 	 int
	
}

type SshKeys struct {
	Host     string
	Port     string
	Username string
	Password string
	Compress bool
	Tunnel   string
	Note     string
	
}

var cf ConfigFile

func main() {

	config := new(Config)
	defaultConfig := new(Config)
	defaultConfig.InjectorSettings = libinject.DefaultConfig
	defaultConfig.SshSettings = libsshclient.DefaultConfig


	header := `
	
░██████╗██╗███╗░░░███╗██████╗░██╗░░░░░███████╗  ████████╗██╗░░░██╗███╗░░██╗███╗░░██╗███████╗██╗░░░░░
██╔════╝██║████╗░████║██╔══██╗██║░░░░░██╔════╝  ╚══██╔══╝██║░░░██║████╗░██║████╗░██║██╔════╝██║░░░░░
╚█████╗░██║██╔████╔██║██████╔╝██║░░░░░█████╗░░  ░░░██║░░░██║░░░██║██╔██╗██║██╔██╗██║█████╗░░██║░░░░░
░╚═══██╗██║██║╚██╔╝██║██╔═══╝░██║░░░░░██╔══╝░░  ░░░██║░░░██║░░░██║██║╚████║██║╚████║██╔══╝░░██║░░░░░
██████╔╝██║██║░╚═╝░██║██║░░░░░███████╗███████╗  ░░░██║░░░╚██████╔╝██║░╚███║██║░╚███║███████╗███████╗
╚═════╝░╚═╝╚═╝░░░░░╚═╝╚═╝░░░░░╚══════╝╚══════╝  ░░░╚═╝░░░░╚═════╝░╚═╝░░╚══╝╚═╝░░╚══╝╚══════╝╚══════╝
`
	author := `▀▄▀▄▀▄ Version 0.1.2 (c) 2021 @pigscanfly ▄▀▄▀▄▀ `
	
	libutils.ClearScreen()
	SetConsoleTitle("Simple Tunnel SSH | @pigscanfly")
	//liblog.LogInfo("Simple Tunnel SSH v0.1.2 | @pigscanfly", "", liblog.Colors["C1"])
	//myFigure := figure.NewFigure("Simple Tunnel SSH", "", true)
	//myFigure.Print()
	fmt.Println(header)
	fmt.Println("\t\t\t" + author)
	//liblog.LogInfo("Version 0.1.2 (c) 2021 @pigscanfly\n", "", liblog.Colors["C1"])  
	liblog.LogInfo("-", "", liblog.Colors["C1"])  
	key := "d5b8f3d1c8adceafbd77fb3c991e4cf78d174dd4b71f988692fc31c46eaa242e"

	if _, err := os.Stat("config/enc-profile.sts"); err == nil {
		isLock = true
		contents, err := readFromFile("config/enc-profile.sts")
		if err != nil {
			liblog.LogInfo("Error reading profile.", "", liblog.Colors["R1"])
		} 
		decrypted := decrypt(string(contents), key)
		json.Unmarshal([]byte(decrypted), &cf)
		liblog.LogInfo("Encrypted profile detected. Reading...", "", liblog.Colors["C1"])
		
	}	else if  _, err := os.Stat("config/profile.json"); err == nil {
		contents, err := readFromFile("config/profile.json")
		if err != nil {
			liblog.LogInfo("Error reading profile.", "", liblog.Colors["R1"])
		} 
		json.Unmarshal([]byte(contents), &cf)
		liblog.LogInfo("Normal profile detected. Reading...", "", liblog.Colors["C1"])
		
	} else {
		liblog.LogInfo("No profile found. Generating profile. Rerun the application.", "", liblog.Colors["C1"])
		libutils.JsonReadWrite(libutils.RealPath("config/profile.json"), config, defaultConfig)
	}
	
	Inject := new(libinject.Inject)
	Inject.Config = config.InjectorSettings
	SshClient := new(libsshclient.SshClient)
	SshClient.Config = config.SshSettings

	hostPtr := cf.InjectorSettings.ProxyHost
	portPtr := cf.InjectorSettings.ProxyPort
	payloadPtr := cf.InjectorSettings.Payload
	listenPtr := cf.InjectorSettings.ListenPort
	timerPtr := cf.InjectorSettings.Timer
	notePtr := cf.SshSettings.Note

	if isLock == true {
		liblog.LogInfo("==========================================", "", liblog.Colors["CC"])
		liblog.LogInfo("Injector running on port: ENCRYPTED" , "", liblog.Colors["B2"])
		liblog.LogInfo("Remote proxy running on: ENCRYPTED", "", liblog.Colors["B2"])
		liblog.LogInfo("Payload: ENCRYPTED", "", liblog.Colors["B2"])
		liblog.LogInfo("Timer: " + strconv.Itoa(timerPtr), "", liblog.Colors["B2"])
		liblog.LogInfo("Note: " + notePtr, "", liblog.Colors["B2"])
		liblog.LogInfo("==========================================", "", liblog.Colors["CC"])
	} else {
		liblog.LogInfo("==========================================", "", liblog.Colors["CC"])
		liblog.LogInfo("Injector running on port: " + listenPtr, "", liblog.Colors["B2"])
		liblog.LogInfo("Remote proxy running on: " + hostPtr + ":" + portPtr, "", liblog.Colors["B2"])
		liblog.LogInfo("Payload: " + payloadPtr, "", liblog.Colors["B2"])
		liblog.LogInfo("Timer: " + strconv.Itoa(timerPtr), "", liblog.Colors["B2"])
		liblog.LogInfo("Note: " + notePtr, "", liblog.Colors["B2"])
		liblog.LogInfo("==========================================", "", liblog.Colors["CC"])
	}

	conn, err := net.Listen("tcp", "127.0.0.1:" + listenPtr)
	liblog.LogInfo("Waiting for client to connect", "", liblog.Colors["C1"])

	if err != nil {
		liblog.LogInfo("Unexpected error [4]", "", liblog.Colors["R1"])
	}

	manager := ClientManager{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		payload:    string(payloadPtr),
	}

	
	go manager.start()
	
	for i := 1; i <= 1; i++ {
		SshClient.InjectPort = cf.InjectorSettings.ListenPort
		SshClient.ListenPort = "1080"
		SshClient.Host = cf.SshSettings.Host
		SshClient.Port = cf.SshSettings.Port
		SshClient.Username = cf.SshSettings.Username
		SshClient.Password = cf.SshSettings.Password
		SshClient.Compress = cf.SshSettings.Compress
		SshClient.Verbose = true
		SshClient.Loop = true
		SshClient.SetRegedit()
		go SshClient.Start() 
	}


	for {

		accept, err := conn.Accept()
		if err != nil {
			liblog.LogInfo("Unexpected error. Restart application", "", liblog.Colors["R1"])
		}

		proxy, err := net.Dial("tcp", hostPtr + ":" + portPtr)
		if err != nil {
			liblog.LogInfo("Unable to connect to the remote proxy server", "", liblog.Colors["R1"])
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
			liblog.LogInfo("Timer Reconnect Initiated", "", liblog.Colors["P1"])
		}

		manager.register <- client
		go manager.handleConnection(client)

	}

}
