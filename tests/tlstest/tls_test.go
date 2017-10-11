package tlstest

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestTLS(t *testing.T) {
	prikey, err := rsa.GenerateKey(rand.Reader, 2048)
	checkError(err)
	t.Logf("private key generated: %d", prikey.D.BitLen())
	pubkey := prikey.Public().(*rsa.PublicKey)
	t.Logf("public key generated: %d", pubkey.N.BitLen())

}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
