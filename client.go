package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"project/client/src/errorchecker"
	"project/client/src/models"
	"project/client/src/utils"
	"strconv"

	"github.com/howeyc/gopass"
)

var origin = "http://localhost:8080"
var currentUser models.User

func printTitle() {
	fmt.Println()
	fmt.Println()
	fmt.Println("Securechat")
	fmt.Println("___________________")
	fmt.Println()
}

func actionsMenu() {
	printTitle()
	fmt.Println("1. Nuevo Chat")
	fmt.Println("2. Ver Chats")
	fmt.Println("q. Cerrar sesión")
}

func mainMenu() {
	printTitle()
	fmt.Println("1. Iniciar Sesión")
	fmt.Println("2. Registarse")
	fmt.Println("q. salir")
}

func bye() {
	fmt.Println("\nVuelve pronto!!")
}

func client(user models.User) {
	fmt.Println("Bienvenido " + user.Username + "!!")
	currentUser = user
	var c string
	for c != "q" {
		actionsMenu()
		fmt.Scanf("%s", &c)
		switch {
		case c == "1":
			fmt.Println("¿Con quién quieres hablar?")
			var peerUsername string
			fmt.Scanf("%s", &peerUsername)
			searchedUsers, _ := models.SearchUser(peerUsername)
			if len(searchedUsers) == 1 {
				startChat(searchedUsers[0])
			} else {
				selection := showUsers(searchedUsers)

				fmt.Println("selection", selection)
				startChat(searchedUsers[selection])
			}

		case c == "2":
			fmt.Println("Listado de chats abiertos")
		}
	}
}

func showUsers(users []models.User) int {
	var userSelected string
	for index, user := range users {
		fmt.Println(index, user.Username)
	}
	fmt.Println("Seleccciona el usuario:")
	fmt.Scanf("%s", &userSelected)
	selection, _ := strconv.Atoi(userSelected)
	return selection
}

func startChat(receiver models.User) {
	word := utils.RandomKey(16)

	receiverPubKey, _ := base64.StdEncoding.DecodeString(receiver.PubKey)
	receiverKey := utils.Myaes(word, receiverPubKey[:32])
	senderPubKey, _ := base64.StdEncoding.DecodeString(currentUser.PubKey)
	senderKey := utils.Myaes(word, senderPubKey[:32])

	res, _ := http.PostForm(origin+"/new_chat", url.Values{
		"sender":      {currentUser.Username},
		"senderkey":   {base64.StdEncoding.EncodeToString(senderKey)},
		"receiver":    {receiver.Username},
		"receiverkey": {base64.StdEncoding.EncodeToString(receiverKey)}})
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))

	conn, err := net.Dial("tcp", "localhost:1337") // llamamos al servidor
	if err != nil {
		fmt.Println("ERROR", err)
	}
	fmt.Println("antes close")
	defer conn.Close() // es importante cerrar la conexión al finalizar
	fmt.Println("despues close")
	fmt.Println("conectado a ", conn.RemoteAddr())

	keyscan := bufio.NewScanner(os.Stdin) // scanner para la entrada estándar (teclado)
	netscan := bufio.NewScanner(conn)     // scanner para la conexión (datos desde el servidor)

	for keyscan.Scan() { // escaneamos la entrada
		fmt.Fprintln(conn, keyscan.Text())         // enviamos la entrada al servidor
		netscan.Scan()                             // escaneamos la conexión
		fmt.Println("servidor: " + netscan.Text()) // mostramos mensaje desde el servidor
	}
}

func registerMenu() {
	var username string
	fmt.Println("Nombre de usuario")
	fmt.Scan(&username)
	fmt.Println("Contraseña")
	pass, _ := gopass.GetPasswd()
	user := models.RegisterUser(username, pass)
	user.Print()
	if user.Username != "" {
		client(user)
	}
}

func doLogin(username string, password []byte) models.User {
	var user models.User
	hashedPassword, _ := utils.Hash(password)
	postURL := origin + "/login"
	parameters := url.Values{"username": {username}, "pass": {utils.Encode64(hashedPassword)}}
	res, err := http.PostForm(postURL, parameters)
	if !errorchecker.Check("ERROR post", err) {
		body, err := ioutil.ReadAll(res.Body)
		if !errorchecker.Check("ERROR read body", err) {
			json.Unmarshal(body, &user)
			res.Body.Close()
		}
	}
	return user
}

func loginMenu() {
	var username string
	fmt.Println("Nombre de usuario")
	fmt.Scan(&username)
	fmt.Println("Contraseña")
	password, err := gopass.GetPasswd()
	if !errorchecker.Check("ERROR contraseña", err) {
		user := doLogin(username, password)
		if user.Validate() {
			client(user)
		}
	}
}

func main() {
	fmt.Println("Iniciando cliente...")
	var option string
	for option != "q" {
		mainMenu()
		fmt.Scanf("%s", &option)
		switch {
		case option == "1":
			loginMenu()
		case option == "2":
			registerMenu()
		}
	}
	bye()
}

//USING Web Sockets

//ws, err := websocket.Dial(url, "", origin)
// if err != nil {
// 	log.Fatal(err)
// }
