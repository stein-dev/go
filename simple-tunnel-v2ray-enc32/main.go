package main

import (
	"syscall"
	"unsafe"
	"fmt"
	"io/ioutil"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"os"
	"bufio"

	"github.com/stein/simple-tunnel-v2ray-enc/libutils"
	"github.com/stein/simple-tunnel-v2ray-enc/libv2ray"
)

var (
	InterruptHandler = new(libutils.InterruptHandler)
	FileName = ""
	DirName = ""
)


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

func init() {
	key := "d5b8f3d1c8adceafbd77fb3c991e4cf78d174dd4b71f988692fc31c46eaa242e"
	if _, err := os.Stat("config/enc-config.stv"); err == nil {
		
		contents, err := readFromFile("config/enc-config.stv")
		if err != nil {
			fmt.Println("Error reading config.")
			return
		} 
		decrypted := decrypt(string(contents), key)
		//fmt.Println(decrypted)
		
		DirName = os.TempDir() + "\\FBC034EF-12ED-XZXZ-ADDD-09F6E90A2D28"
		//fmt.Println(os.TempDir())
		//fmt.Println(DirName)

		if _, err := os.Stat(DirName); !os.IsNotExist(err) {
			os.RemoveAll(DirName)
		}
		_, err2 := os.Stat(DirName)
 
		if os.IsNotExist(err2) {
			errDir := os.MkdirAll(DirName, 0755)
			if errDir != nil {
				//log.Fatal(err2)
				os.Exit(1)
			}
		}

		file, err := ioutil.TempFile(DirName, "09F6E90A2D28")
		if err != nil {
			//log.Fatal(err)
			os.Exit(1)
		}

		defer os.RemoveAll(file.Name())
		
		FileName = file.Name()
		fmt.Println(FileName)
		writeToFile(string(decrypted), file.Name())

	} else {
		fmt.Println("No config found.")
		return
	}
	InterruptHandler.Handle = func() {
		os.RemoveAll(FileName)
		libv2ray.Stop()
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

	V2RayClient := new(libv2ray.V2RayClient)
	V2RayClient.ConfigName = FileName
	V2RayClient.Start()

	InterruptHandler.Wait(FileName)
}
