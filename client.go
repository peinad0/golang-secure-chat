package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"project/client/src/constants"
	"project/client/src/errorchecker"
	"project/client/src/models"
	"project/client/src/utils"
	"strconv"

	"github.com/howeyc/gopass"
)

var origin = constants.ServerOrigin
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

func client() {
	fmt.Println("Bienvenido " + currentUser.Username + "!")
	var c string
	for c != "q" {
		actionsMenu()
		fmt.Scanf("%s", &c)
		switch {
		case c == "1":
			fmt.Println("¿Con quién quieres hablar?")
			var peerUsername string
			fmt.Scanf("%s", &peerUsername)

			searchedUsers, _ := models.SearchUsers(peerUsername)
			selection := showUsers(searchedUsers)
			if selection != -1 {
				models.StartChat(currentUser, searchedUsers[selection])
			}
		case c == "2":
			fmt.Println("Listado de chats abiertos:")
			searchedChats, _ := models.GetChats(currentUser)
			selection := showChats(searchedChats)
			if selection != -1 {
				models.OpenChat(searchedChats[selection], currentUser)
			}

		}
	}

	postURL := origin + "/logout"
	parameters := url.Values{"username": {currentUser.Username}}
	_, err := http.PostForm(postURL, parameters)
	if !errorchecker.Check("ERROR logout", err) {
		fmt.Println("Logout successful")
	}
}

func showUsers(users []models.User) int {
	var userSelected string
	if len(users) > 0 {
		for index, user := range users {
			fmt.Println(index, user.Username)
		}
		fmt.Scanf("%s", &userSelected)
		selection, err := strconv.Atoi(userSelected)
		if err == nil && selection >= 0 && selection < len(users) {
			return selection
		}
		fmt.Println("Error seleccionando usuario.")
		return -1
	}
	fmt.Println("No se encontraron usuarios.")
	return -1
}

func showChats(chats []models.Chat) int {
	var chatSelected string
	if len(chats) > 0 {
		for index, chat := range chats {
			fmt.Println(index, chat.Name)
		}
		fmt.Println("Seleccciona el chat:")
		fmt.Scanf("%s", &chatSelected)
		selection, err := strconv.Atoi(chatSelected)
		if err == nil && selection >= 0 && selection < len(chats) {
			return selection
		}
		fmt.Println("Error seleccionando chat.")
		return -1
	}
	fmt.Println("No se encontraron chats.")
	return -1
}

func doLogin(username string, password []byte) models.User {
	var user models.User
	passwordSlice, encrypterSlice := utils.Hash(password)
	postURL := origin + "/login"
	parameters := url.Values{"username": {username}, "pass": {utils.Encode64(passwordSlice)}}
	res, err := http.PostForm(postURL, parameters)
	if !errorchecker.Check("ERROR post", err) {
		body, err := ioutil.ReadAll(res.Body)
		if !errorchecker.Check("ERROR read body", err) {
			json.Unmarshal(body, &user)
			res.Body.Close()
			fmt.Println(encrypterSlice)
			//user.Decode(encrypterSlice)
		}
	}
	return user
}

func registerMenu() {
	var username string
	fmt.Println("Nombre de usuario")
	fmt.Scan(&username)
	fmt.Println("Contraseña")
	pass, _ := gopass.GetPasswd()
	user := models.RegisterUser(username, pass)
	user.Print()
	if user.Validate() {
		currentUser = user
		client()
	}
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
			currentUser = user
			client()
		}
	}
}

func main() {
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
