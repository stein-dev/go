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
	"runtime"
	"os/exec"
)

func main() {

	// bytes := make([]byte, 32) //generate a random 32 byte key for AES-256
	// if _, err := rand.Read(bytes); err != nil {
	// 	panic(err.Error())
	// }

	key := "d5b8f3d1c8adceafbd77fb3c991e4cf78d174dd4b71f988692fc31c46eaa242e" //encode key in bytes to string and keep as secret, put in a vault
	//fmt.Printf("key to encrypt/decrypt : %s\n", key)

	// encrypted := encrypt("Hello Encrypt", key)
	// fmt.Printf("encrypted : %s\n", encrypted)

	// decrypted := decrypt(encrypted, key)
	// fmt.Printf("decrypted : %s\n", decrypted)

	

	ClearScreen()
	fmt.Printf("OpenVPN Config Generator\n")
	fmt.Printf("[1] Encrypt Config\n")
	fmt.Printf("[2] Decrypt Config\n")
	fmt.Printf("Choose Number: ")
	choice := readline()

	switch choice {
	case "1":
		fmt.Printf("==============================\n")
		fmt.Printf("File to be encrypted: ")
		enFile := readline()
		fmt.Printf("Encrypting file [%s] in progress...", enFile)
		contents, err := readFromFile(enFile)
		if err != nil {
			fmt.Println("\nFile not found")
		} else {
			fmt.Printf("\nContents Before Encryption: \n%s", contents)
			encrypted := encrypt(string(contents), key)
			fmt.Printf("\nContents After Encryption: \n%s", encrypted)
			fmt.Printf("\nWriting contents to file [enc-%s]...", enFile)
			writeToFile(string(encrypted), "enc-"+ enFile)
			fmt.Printf("\nDone!\n")
		}
		
	case "2":
		fmt.Printf("==============================\n")
		fmt.Printf("File to be decrypted: ")
		enFile := readline()
		fmt.Printf("Decrypting file [%s] in progress...", enFile)
		contents, err := readFromFile(enFile)
		if err != nil {
			fmt.Println("\nFile not found")
		} else {
			fmt.Printf("\nContents Before Decryption: \n%s", contents)
			decrypted := decrypt(string(contents), key)
			fmt.Printf("\nContents After Decryption: \n%s", decrypted)
			//fmt.Printf("\nWriting contents to file [dec-%s]...", enFile)
			//writeToFile(decrypted, "dec-"+ enFile)
			//fmt.Printf("\nDone!\n")
		}
	default:
		fmt.Printf("Invalid choice. Now Exiting... ")
		os.Exit(1)
	}
}

func ClearScreen() {
	switch runtime.GOOS {
	case "linux", "android":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
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
