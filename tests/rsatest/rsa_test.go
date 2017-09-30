package rsatest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestRSA(t *testing.T) {
	pemFile, err := os.Open("h39.pem")
	checkError(err)
	defer pemFile.Close()
	pemData, err := ioutil.ReadAll(pemFile)
	checkError(err)
	privateBlock, _ := pem.Decode(pemData)
	priKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	checkError(err)

	pubFile, err := os.Open("h39.pub")
	checkError(err)
	defer pubFile.Close()
	pubData, err := ioutil.ReadAll(pubFile)
	checkError(err)
	pubBlock, _ := pem.Decode(pubData)
	pubKeyValue, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	checkError(err)
	pubKey := pubKeyValue.(*rsa.PublicKey)

	var plainText []byte
	for i := 0; i < 256-50; i++ {
		plainText = append(plainText, 'A')
	}

	encryptText, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, pubKey, plainText, nil)
	checkError(err)
	ioutil.WriteFile("encrypt.txt", encryptText, 0644)

	decryptText, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, priKey, encryptText, nil)
	checkError(err)

	fmt.Printf("Decrypt: %s", string(decryptText))
	if string(decryptText) != string(plainText) {
		t.Fatalf("decrypt text is wrong")
	}
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
