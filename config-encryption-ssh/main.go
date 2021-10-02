package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"bufio"
	"os/exec"
	"syscall"
	"unsafe"
)

func main() {

	// bytes := make([]byte, 32) //generate a random 32 byte key for AES-256
	// if _, err := rand.Read(bytes); err != nil {
	// 	panic(err.Error())
	// }

	SetConsoleTitle("Config Encryptor | @pigscanfly")
	key := "d5b8f3d1c8adceafbd77fb3c991e4cf78d174dd4b71f988692fc31c46eaa242e" //encode key in bytes to string and keep as secret, put in a vault
	//fmt.Printf("key to encrypt/decrypt : %s\n", key)

	// encrypted := encrypt("Hello Encrypt", key)
	// fmt.Printf("encrypted : %s\n", encrypted)

	// decrypted := decrypt(encrypted, key)
	// fmt.Printf("decrypted : %s\n", decrypted)

	

	ClearScreen()
	fmt.Printf("Simple Tunnel SSH Config Encryptor | @pigscanfly\n\n")
	fmt.Printf("[*] This program encrypts the profile.json config file of Simple Tunnel SSH.\n")
	fmt.Printf("[*] This is useful if you want to share the config to others without\n")
	fmt.Printf("    sharing your one of a kind payload and ssh credentials.\n")
	fmt.Printf("[*] The config(profile.json) must be in the same directory as the program.\n")
	fmt.Printf("[*] This will generate a file \"enc-profile.sts\". Place this config inside config folder to use.\n")
	fmt.Printf("\nEnter the filename(profile.json) to be encrypted: ")
	enFile := readline()
	fmt.Printf("Encrypting file [%s] in progress...", enFile)
	contents, err := readFromFile(enFile)
	if err != nil {
		fmt.Println("\nFile not found")
	} else {
		//fmt.Printf("\nContents Before Encryption: \n%s", contents)
		encrypted := encrypt(string(contents), key)
		//fmt.Printf("\nContents After Encryption: \n%s", encrypted)
		fmt.Printf("\nWriting contents to file [enc-profile.sts]...")
		//writeToFile(string(encrypted), "enc-profile.sts")
		file, err := os.Create("enc-profile.sts")

		if err != nil {
			return
		}
		defer file.Close()

		file.WriteString(string(encrypted))
		fmt.Printf("\nDone!\n")
		fmt.Printf("\nPress any key to exit...!\n")
		readline()
	}
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

func ClearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
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
