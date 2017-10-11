package rsatest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestRSA(t *testing.T) {
	pemData, err := ioutil.ReadFile("../../goworld.pem")
	checkError(err)
	privateBlock, _ := pem.Decode(pemData)
	priKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	checkError(err)

	pubData, err := ioutil.ReadFile("../../goworld.pub")
	checkError(err)
	pubBlock, _ := pem.Decode(pubData)

	pubKeyValue, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	checkError(err)
	pubKey := pubKeyValue.(*rsa.PublicKey)

	var plainText []byte
	for i := 0; i < 50; i++ {
		plainText = append(plainText, 'A')
	}

	k := (pubKey.N.BitLen() + 7) / 8
	maxPlainSize := k - 2*sha1.New().Size() - 2
	fmt.Printf("Public key K = %d, max plain text size = %d\n", k, maxPlainSize)

	hash := sha1.New()
	encryptText, err := rsa.EncryptOAEP(hash, rand.Reader, pubKey, plainText, nil)
	checkError(err)
	fmt.Printf("Encrypt text: (%d)%s\n", len(encryptText), encryptText)

	decryptText, err := rsa.DecryptOAEP(hash, rand.Reader, priKey, encryptText, nil)
	checkError(err)

	fmt.Printf("Decrypt: %s\n", string(decryptText))
	if string(decryptText) != string(plainText) {
		t.Fatalf("decrypt text is wrong")
	}

}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
