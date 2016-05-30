package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"project/client/src/errorchecker"
	"project/client/src/models"
	"project/client/src/utils"

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

func menu() {
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
	var c string
	for c != "q" {
		menu()
		fmt.Scanf("%s", &c)
		switch {
		case c == "1":
			searchedUser, _ := models.SearchUser(user.Username)
			if len(searchedUser) == 1 {
				startChat(searchedUser[0].Username, searchedUser[0].PubKey)
			} else {
			}

		case c == "2":
			fmt.Println("Listado de chats abiertos")
		}
	}
}

func startChat(username, pubKeystr string) {
	fmt.Println(username)

	word := utils.RandomKey(16)
	pubKey, _ := base64.StdEncoding.DecodeString(pubKeystr)
	encripted := utils.Myaes(word, pubKey)
	//myEncripted := myaes(word, myPubKey)
	//////
	res, _ := http.PostForm(origin+"/new_chat", url.Values{"username": {username}, "key": {base64.StdEncoding.EncodeToString(encripted)}})
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))
}

func registerMenu() {
	var username string
	fmt.Println("Nombre de usuario")
	fmt.Scan(&username)
	fmt.Println("Contraseña")
	pass, _ := gopass.GetPasswd()
	user := models.RegisterUser(username, pass)
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
