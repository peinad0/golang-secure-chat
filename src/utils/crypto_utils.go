package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"hash"
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
	var privateKey, publicKey []byte
	var errPriv, errPub error

	privateKey, errPriv = json.Marshal(privKey)
	if !errorchecker.Check("ERROR Marshall Priv", errPriv) {
		publicKey, errPub = json.Marshal(privKey.PublicKey)
		if !errorchecker.Check("ERROR Marshall Pub", errPub) {
			privateKey = EncryptAES(Compress(privateKey), key)
		}
	}
	return privateKey, Compress(publicKey)
}

// GetKeys function
func GetKeys() (*rsa.PrivateKey, *rsa.PublicKey) {
	priv := GenerateKeys()
	return priv, &priv.PublicKey
}

// GenerateKeys function
func GenerateKeys() *rsa.PrivateKey {
	privKey, err := rsa.GenerateKey(crand.Reader, 2048)
	if !errorchecker.Check("ERROR generateKeys", err) {
		privKey.Precompute() // acelera el uso con precálculo
	}
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

// EncryptAES función para cifrar (con AES en este caso)
func EncryptAES(data, key []byte) (out []byte) {
	out = make([]byte, len(data)+16)         // reservamos espacio para el IV al principio
	rand.Read(out[:16])                      // generamos el IV
	blk, err := aes.NewCipher(key)           // cifrador en bloque (AES), usa key
	errorchecker.Check("ERROR encrypt", err) // comprobamos el error
	ctr := cipher.NewCTR(blk, out[:16])      // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out[16:], data)         // ciframos los datos
	return
}

// DecryptAES función para descifrar
func DecryptAES(data, key []byte) (out []byte) {
	out = make([]byte, len(data)-16)         // la salida no va a tener el IV
	blk, err := aes.NewCipher(key)           // cifrador en bloque (AES), usa key
	errorchecker.Check("ERROR decrypt", err) // comprobamos el error
	ctr := cipher.NewCTR(blk, data[:16])     // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out, data[16:])         // desciframos (doble cifrado) los datos
	return
}

//EncryptOAEP func
func EncryptOAEP(publicKey *rsa.PublicKey, plainText, label []byte) (encrypted []byte) {
	var err error
	var md5Hash hash.Hash

	md5Hash = md5.New()

	if encrypted, err = rsa.EncryptOAEP(md5Hash, crand.Reader, publicKey, plainText, label); err != nil {
		log.Fatal(err)
	}
	return
}

//DecryptOAEP func
func DecryptOAEP(privateKey *rsa.PrivateKey, encrypted, label []byte) (decrypted []byte) {
	var err error
	var md5Hash hash.Hash

	md5Hash = md5.New()
	if decrypted, err = rsa.DecryptOAEP(md5Hash, crand.Reader, privateKey, encrypted, label); err != nil {
		log.Fatal(err)
	}
	return
}

// RandomKey function returning a byte array of size 'size'
func RandomKey(size int) []byte {
	encripted := make([]byte, size)
	rand.Read(encripted)
	return encripted
}
