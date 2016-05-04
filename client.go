/*

Este programa demuestra una arquitectura cliente servidor sencilla.
El cliente envía líneas desde la entrada estandar y el servidor le devuelve un reconomiento de llegada (acknowledge).
El servidor es concurrente, siendo capaz de manejar múltiples clientes simultáneamente.
Las entradas se procesan mediante un scanner (bufio).

ejemplos de uso:

go run cnx.go srv

go run cnx.go cli

*/

package main

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/howeyc/gopass"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var origin = "http://localhost:8080"

// User structure
type User struct {
	Username  string `bson:"name"`
	Password  string `bson:"password"`
	Salt      string `bson:"salt"`
	PubKey    string `bson:"pubkey"`
	PrivKey   string `bson:"privkey"`
	CipherMsg string `bson:"ciphermsg"`
}

// función para comprobar errores (ahorra escritura)
func chk(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	fmt.Println("Iniciando cliente...")

	//ws, err := websocket.Dial(url, "", origin)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	var c string
	for c != "q" {
		loginMenu()

		fmt.Scanf("%s", &c)
		switch {
		case c == "1":
			login()
		case c == "2":
			register()
		}
	}

	bye()
}

func client(user string /*ws *websocket.Conn*/) {
	fmt.Println("Bienvenido " + user + "!!")
	var c string
	for c != "q" {
		menu()
		fmt.Scanf("%s", &c)
		switch {
		case c == "1":
			fmt.Println("Listado usuarios conectados")
		case c == "2":
			fmt.Println("Listado de chats abiertos")
		}
	}
}

func menu() {
	fmt.Println()
	fmt.Println()
	fmt.Println("Securechat")
	fmt.Println("___________________")
	fmt.Println()
	fmt.Println("1. Nuevo Chat")
	fmt.Println("2. Ver Chats")
	fmt.Println("q. Cerrar sesión")
}

func loginMenu() {
	fmt.Println()
	fmt.Println()
	fmt.Println("Securechat")
	fmt.Println("___________________")
	fmt.Println()
	fmt.Println("1. Iniciar Sesión")
	fmt.Println("2. Registarse")
	fmt.Println("q. salir")
}

func bye() {
	fmt.Println()
	fmt.Println("Adios!!")
}

func register() {
	var username string
	fmt.Println("Nombre de usuario")
	fmt.Scan(&username)
	fmt.Println("Contraseña")
	pass, _ := gopass.GetPasswd()

	passEnc, keyEnc := hash(pass)

	pub, priv := getKeys(keyEnc)

	encodedPass := base64.StdEncoding.EncodeToString(passEnc)
	encodedPub := base64.StdEncoding.EncodeToString(pub)
	encodedPriv := base64.StdEncoding.EncodeToString(priv)

	//
	// #DBUG

	publicKey := rsa.PublicKey{}

	json.Unmarshal(pub, &publicKey)

	var label []byte

	test, err := rsa.EncryptOAEP(sha512.New(), rand.Reader, &publicKey, []byte(username), label)

	if err != nil {
		fmt.Println("ERROR OAEP", err)
	}

	encodedTest := base64.StdEncoding.EncodeToString(test)

	//
	//

	res, err := http.PostForm(origin+"/register", url.Values{"username": {username}, "pass": {encodedPass}, "pub": {encodedPub}, "priv": {encodedPriv}, "msg": {encodedTest}})

	if err != nil {
		fmt.Println("Error en POST")
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error read", err)
	}

	registered := string(body)

	res.Body.Close()

	if registered == "true" {
		client(username)
	}
}

func getKeys(key []byte) (public, private []byte) {

	priv := generateKeys()

	private, public = cipherKeys(priv, key)

	return
}

func generateKeys() *rsa.PrivateKey {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		fmt.Println("ERROR")
	}

	privKey.Precompute() // acelera el uso con precálculo

	return privKey
}

func cipherKeys(privKey *rsa.PrivateKey, key []byte) (priv, pub []byte) {
	block, err := aes.NewCipher(key)

	if err != nil {
		fmt.Println("ERROR AES")
	}

	priv, err = json.Marshal(privKey)

	if err != nil {
		fmt.Println("ERROR MARSHAL PRIV")
	}

	pub, err = json.Marshal(privKey.PublicKey)

	if err != nil {
		fmt.Println("ERROR MARSHAL PUB")
	}

	block.Encrypt(priv, priv)

	return
}

func hash(pass []byte) (keyPass, keyEnc []byte) {
	hash := sha512.Sum512(pass)

	keyPass = hash[:32]
	keyEnc = hash[32:]

	return
}

func login() {
	var username string
	fmt.Println("Nombre de usuario")
	fmt.Scan(&username)
	fmt.Println("Contraseña")
	pass, _ := gopass.GetPasswd()

	passEnc, keyEnc := hash(pass)

	encodedString := base64.StdEncoding.EncodeToString(passEnc)

	res, err := http.PostForm(origin+"/login", url.Values{"username": {username}, "pass": {encodedString}})

	if err != nil {
		fmt.Println("Error en POST")
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error read", err)
	}

	var user User

	json.Unmarshal(body, &user)

	res.Body.Close()

	private, _ := base64.StdEncoding.DecodeString(user.PrivKey)
	decodedMsg, _ := base64.StdEncoding.DecodeString(user.CipherMsg)

	msgTest := descifrar(decodedMsg, keyEnc, private)

	fmt.Println("USUARIO", string(msgTest))

	if user.Username != "" {
		client(user.Username)
	}
}

func descifrar(msg, pass, private []byte) []byte {
	block, err := aes.NewCipher(pass)

	if err != nil {
		fmt.Println("ERROR aes", err)
	}

	// modo de operacion newCTR o new OFB
	block.Decrypt(private, private)

	p := rsa.PrivateKey{}

	err = json.Unmarshal(private, &p)

	var label []byte
	mensaje, err := rsa.DecryptOAEP(sha512.New(), rand.Reader, &p, msg, label)

	if err != nil {
		fmt.Println("ERROR DecryptOAEP", err)
	}

	return mensaje
}

func send(ws *websocket.Conn, m string) {

	message := []byte(m)
	_, err := ws.Write(message)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Send: %s\n", message)

	var msg = make([]byte, 512)
	_, err = ws.Read(msg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Receive: %s\n", msg)
}
