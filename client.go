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
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"github.com/howeyc/gopass"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var origin = "http://localhost:8080"

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
	var user string
	fmt.Println("Nombre de usuario")
	fmt.Scan(&user)
	fmt.Println("Contraseña")
	pass, _ := gopass.GetPasswd()

	passEnc, _ := hash(pass)

	encodedString := base64.StdEncoding.EncodeToString(passEnc)

	res, err := http.PostForm(origin+"/register", url.Values{"user": {user}, "pass": {encodedString}})

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
		client(user)
	}
}

func hash(pass []byte) (keyPass, keyEnc []byte) {
	hash := sha512.Sum512(pass)

	keyPass = hash[:32]
	keyEnc = hash[32:]

	return
}

func login() {
	var user string
	fmt.Println("Nombre de usuario")
	fmt.Scan(&user)
	fmt.Println("Contraseña")
	pass, _ := gopass.GetPasswd()

	passEnc, _ := hash(pass)

	encodedString := base64.StdEncoding.EncodeToString(passEnc)

	res, err := http.PostForm(origin+"/login", url.Values{"user": {user}, "pass": {encodedString}})

	if err != nil {
		fmt.Println("Error en POST")
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error read", err)
	}

	logged := string(body)

	res.Body.Close()

	if logged == "true" {
		client(user)
	}
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
