package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
			selection := showUsers(searchedUsers)

			models.StartChat(currentUser, searchedUsers[selection])
		case c == "2":
			fmt.Println("Listado de chats abiertos")
			searchedChats, _ := models.GetChats(user)
			selection := showChats(searchedChats)

			models.OpenChat(searchedChats[selection], user)
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

func showChats(chats []models.Chat) int {
	var chatSelected string
	for index, chat := range chats {
		fmt.Println(index, chat.Name)
	}
	fmt.Println("Seleccciona el chat:")
	fmt.Scanf("%s", &chatSelected)
	selection, _ := strconv.Atoi(chatSelected)
	return selection
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
