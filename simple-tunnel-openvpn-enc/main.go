package main

import (
	"io"
	"net"
	"strings"
	"time"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"bufio"
	"encoding/json"
	"syscall"
	"unsafe"

	"github.com/stein/simple-tunnel-openvpn-enc/libinject"
	"github.com/stein/simple-tunnel-openvpn-enc/liblog"
	"github.com/stein/simple-tunnel-openvpn-enc/libutils"
	"github.com/stein/simple-tunnel-openvpn-enc/libopenvpn"

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

type Config struct {
	InjectorSettings *libinject.Config
}


func (manager *ClientManager) start() {
	for {
		select {
		case connection := <-manager.register:
			manager.clients[connection] = true
			//log.Println("Client Connected:", connection.socket.RemoteAddr())
			liblog.LogInfo("Client Connected", "", Colors["C1"])
		case connection := <-manager.unregister:
			delete(manager.clients, connection)
			//log.Println("Client Disconnected:", connection.socket.RemoteAddr())
			liblog.LogInfo("Client Disconnected", "", Colors["R1"])
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
		} else {
			parsed = strings.ReplaceAll(parsed, "[protocol]", connProto+"\r\n"+
				"Authorization: Basic "+manager.auth)
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
						liblog.LogInfo("Unexpected error [1]", "", Colors["R1"])
						break
					}
					client.proxy.connected = true
					

				} else {
					nw, err := client.proxy.socket.Write(buf[0:nr])
					if err != nil {
						//log.Println(err)
						liblog.LogInfo("Unexpected error [2]", "", Colors["R1"])
						break
					}

					if nr != nw {
						//log.Println(err)
						liblog.LogInfo("Unexpected error [3]", "", Colors["R1"])
						break
					}
				}
			}

			if err != nil {
				if err != io.EOF {
					//log.Println(err)
					//log.Println("Server unexpectedly closed network connection")
					liblog.LogInfo("Server unexpectedly closed network connection", "", Colors["R1"])
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
		liblog.LogInfo("Injecting Payload", "", Colors["P1"])
		for {
			nr, err := r.Read(buf)
			data := ""
			data += string(buf[:nr])
			
			if nr > 0 {
				nw, err := client.socket.Write(buf[0:nr])

				if strings.Contains(strings.Split(data, "\r\n")[0], "HTTP/1.1 200 Connection established") {
					//log.Println("HTTP/1.1 200 Connection established")
					liblog.LogInfo("HTTP/1.1 200 Connection established", "", Colors["G1"])
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


func encrypt(stringToEncrypt string, keyString string) (encryptedString string) {

	//Since the key is in string, we need to convert decode it to bytes
	key, _ := hex.DecodeString(keyString)
	plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext)
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

type InjectorKeys struct {
	ListenPort string
	Payload string
	ProxyHost string
	ProxyPort string
	Timer int
	FileName string
	AuthFileName string
	OpenVPNConfig string
}

type ConfigFile struct {
	InjectorSettings InjectorKeys
  }

var cf ConfigFile  

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

	SetConsoleTitle("Simple Tunnel OpenVPN - v0.1.1 | @pigscanfly")
	libutils.ClearScreen()
	key := "d5b8f3d1c8adceafbd77fb3c991e4cf78d174dd4b71f988692fc31c46eaa242e"

	contents, err := readFromFile("config/enc-profile.json")
	if err != nil {
		fmt.Println("\nFile not found")
	} else {
		decrypted := decrypt(string(contents), key)
		//fmt.Printf("\nContents After Decryption: \n%s", decrypted)
		
		json.Unmarshal([]byte(decrypted), &cf)
		//fmt.Println(cf.InjectorSettings.Payload)
		// fmt.Println(cf.InjectorSettings.FileName)
		
	}



	liblog.LogInfo("Simple Tunnel OpenVPN | @pigscanfly", "", Colors["C1"])
	//liblog.LogInfo("This exclusive release is for UDP Team only. Begone spy and snipers!", "", Colors["C1"])
	
	config := new(Config)
	defaultConfig := new(Config)
	defaultConfig.InjectorSettings = libinject.DefaultConfig


	//libutils.JsonReadWrite(libutils.RealPath("profile.json"), config, defaultConfig)

	Inject := new(libinject.Inject)
	Inject.Config = config.InjectorSettings
	Openvpn := new(libopenvpn.Openvpn)

	hostPtr := "18.141.217.149"
	portPtr := "7778"
	payloadPtr := "HTTP//1.1 200 [lf]Host: www.xbox.com [lf][lf][lf]"
	listenPtr := "9292"
	//authPtr := Inject.Config.Username + ":" + Inject.Config.Password
	timerPtr := 55


	// if *hostPtr == "" || *payloadPtr == "" {
	// 	liblog.LogInfo("No arguments supplied. Exiting", "INFO", Colors["R1"])
	// 	os.Exit(1)
	// }

	// sDec, _ := base64.StdEncoding.DecodeString(payloadPtr)
	// var pl = strings.TrimSpace(string(sDec))

	liblog.LogInfo("=================================", "", Colors["C1"])
	//liblog.LogInfo("Injector running on port: LOCKED", "", Colors["C1"])
	//liblog.LogInfo("Remote proxy running on: LOCKED", "", Colors["C1"])
	//liblog.LogInfo("Payload: LOCKED", "", Colors["C1"])
	//liblog.LogInfo("Expiration: Until the end of dawn", "", Colors["C1"])
	liblog.LogInfo("Note: @djdoolky76's server", "", Colors["C1"])
	liblog.LogInfo("=================================", "", Colors["C1"])

	conn, err := net.Listen("tcp", "127.0.0.1:" + listenPtr)
	liblog.LogInfo("Waiting for client to connect", "", Colors["C1"])

	if err != nil {
		//log.Println(err)
		//liblog.LogInfo(string(err), "INFO", Colors["R1"])
		liblog.LogInfo("Unexpected error [4]", "", Colors["R1"])
		
	}

	manager := ClientManager{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		payload:    string(payloadPtr),
		//auth:       base64.StdEncoding.EncodeToString([]byte(authPtr)),
	}

	go manager.start()
	Openvpn.ProxyHost = "18.141.217.149"
	Openvpn.InjectPort = "9292"
	Openvpn.FileName = "config/thedj.ovpn"
	Openvpn.AuthFileName = "config/thedj.auth"
	go Openvpn.Start()

	for {

		accept, err := conn.Accept()
		if err != nil {
			//log.Println("test5")
			liblog.LogInfo("Unexpected error. Restart application", "", Colors["R1"])
		}

		proxy, err := net.Dial("tcp", hostPtr + ":" + portPtr)
		if err != nil {
			liblog.LogInfo("Unable to connect to the remote proxy server", "", Colors["R1"])
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
			liblog.LogInfo("Timer Reconnect Initiated", "", Colors["P1"])
		}

		

		manager.register <- client
		go manager.handleConnection(client)

	}

}
