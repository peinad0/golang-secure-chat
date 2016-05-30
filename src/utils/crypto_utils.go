package utils

import (
	"crypto/aes"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"project/client/src/errorchecker"

	"golang.org/x/crypto/scrypt"
)

// ScryptHash asd
func ScryptHash(word, salt []byte) ([]byte, error) {
	return scrypt.Key(word, salt, 16384, 8, 1, 32)
}

// HashWithRandomSalt asd
func HashWithRandomSalt(pass []byte) ([]byte, []byte) {
	salt := make([]byte, 16)
	rand.Read(salt)
	dk, err := ScryptHash(pass, salt)
	if err != nil {
		log.Println("ERROR SCRYPT", err)
	}
	return dk, salt
}

// Hash receives a byte array, generates a hashed result and splits it into
// two different byte arrays
func Hash(password []byte) ([]byte, []byte) {
	hashedPassword := sha512.Sum512(password)
	passwordSlice := hashedPassword[:32]
	encripterSlice := hashedPassword[32:]
	return passwordSlice, encripterSlice
}

// CipherKeys function receiving a rsa.Private key and a key in order to
// cipher the private key with AES using that key. It also returns the
// public key.
func CipherKeys(privKey *rsa.PrivateKey, key []byte) ([]byte, []byte) {
	block, err := aes.NewCipher(key)
	var privateKey, publicKey []byte
	var errPriv, errPub error
	if !errorchecker.Check("ERROR AES 49", err) {
		privateKey, errPriv = json.Marshal(privKey)
		if !errorchecker.Check("ERROR Marshall Priv", errPriv) {
			publicKey, errPub = json.Marshal(privKey.Public())
			if !errorchecker.Check("ERROR Marshall Pub", errPub) {
				block.Encrypt(privateKey, publicKey)
			}
		}
	}
	return privateKey, publicKey
}

// GetKeys function
func GetKeys(key []byte) (public, private []byte) {
	priv := GenerateKeys()
	private, public = CipherKeys(priv, key)
	return
}

// GenerateKeys function
func GenerateKeys() *rsa.PrivateKey {
	privKey, err := rsa.GenerateKey(crand.Reader, 2048)
	if err != nil {
		fmt.Println("ERROR generatekeys 72")
	}
	privKey.Precompute() // acelera el uso con precálculo
	return privKey
}

// Descifrar function
func Descifrar(msg, pass, private []byte) []byte {
	block, err := aes.NewCipher(pass)
	if err != nil {
		fmt.Println("ERROR aes", err)
	}
	// modo de operacion newCTR o new OFB
	block.Decrypt(private, private)
	p := rsa.PrivateKey{}
	err = json.Unmarshal(private, &p)
	var label []byte
	mensaje, err := rsa.DecryptOAEP(sha512.New(), crand.Reader, &p, msg, label)
	if err != nil {
		fmt.Println("ERROR DecryptOAEP", err)
	}
	return mensaje
}

// RandomKey function
func RandomKey(size int) []byte {
	encripted := make([]byte, size)
	rand.Read(encripted)
	return encripted
}

// Myaes function
func Myaes(word, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("ERROR AES 107")
	}
	block.Encrypt(word, word)
	return word
}
