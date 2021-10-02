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
	fmt.Println("*********************************")
	fmt.Println("*                               *")
	fmt.Println("* Simple Tunnel V2RAY Encryptor *")
	fmt.Println("*                               *")
	fmt.Println("*********************************")
	fmt.Println("")
	fmt.Println("Note:")
	fmt.Println("      Execute this program in the same location of the config that you're going to encrypt")
	fmt.Println("      This will generate a file called enc-config.stv")
	fmt.Println("      Place this file inside config folder")
	fmt.Println("")
	fmt.Println("")
	fmt.Print("Enter file name to encrypt: ")
	enFile := readline()
	fmt.Printf("Encrypting file [%s] in progress...", enFile)
	contents, err := readFromFile(enFile)
	if err != nil {
		fmt.Println("\nFile not found")
	} else {
		//fmt.Printf("\nContents Before Encryption: \n%s", contents)
		fmt.Println("Encrypting...")
		encrypted := encrypt(string(contents), key)
		//fmt.Printf("\nContents After Encryption: \n%s", encrypted)
		fmt.Printf("\nWriting contents to file [enc-%s]...", enFile)
		writeToFile(string(encrypted), "enc-config.stv")
		fmt.Printf("\nDone!\n")
		readline()
	}
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
