package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"project/client/src/constants"
	"project/client/src/errorchecker"
	"project/client/src/models"
	"project/client/src/utils"
	"strconv"
	"strings"
	"syscall"

	"github.com/howeyc/gopass"
)

var origin = constants.ServerOrigin
var currentUser models.PrivateUser

var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var https = &http.Client{Transport: tr}

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
	fmt.Println("3. Ver Contactos")
	fmt.Println("4. Buscar Usuarios")
	fmt.Println("5. Administrar Grupos")
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
			searchedUsers := models.SearchUsers(peerUsername)
			userIDS, err := showUsers(searchedUsers)
			var users []models.PublicUser
			if !errorchecker.Check("ERROR parseando usuarios", err) {
				for _, id := range userIDS {
					users = append(users, searchedUsers[id])
				}
				models.StartChat(currentUser, users)
			}
			break
		case c == "2":
			fmt.Println("Listado de chats abiertos:")
			searchedChats, _ := models.GetChats(currentUser)
			fmt.Println("aveces")
			selection := showChats(searchedChats)
			if selection != -1 {
				models.OpenChat(searchedChats[selection], currentUser)
			}
			break
		case c == "3":
			fmt.Println("¿Con quién quieres hablar?")
			userIDS, err := showUsers(currentUser.State.Contacts)
			var users []models.PublicUser
			if !errorchecker.Check("ERROR", err) {
				for _, id := range userIDS {
					users = append(users, currentUser.State.Contacts[id])
				}
				models.StartChat(currentUser, users)
			}
			break
		case c == "4":
			fmt.Println("Busca los usuarios para añadir a contactos:")
			var peerUsername string
			fmt.Scanf("%s", &peerUsername)
			searchedUsers := models.SearchUsers(peerUsername)
			userIDS, err := showUsers(searchedUsers)
			var users []models.PublicUser
			if !errorchecker.Check("ERROR parseando usuarios", err) {
				for _, id := range userIDS {
					users = append(users, searchedUsers[id])
				}
				currentUser.AddUsersToContacts(users)
			}
			break
		case c == "5":
			fmt.Println("Elegir chat para administrar:")
			chats := models.GetAdminChats(currentUser.Username)
			selection := showChats(chats)
			if selection != -1 {
				models.AdministrarChat(&currentUser, chats[selection])
			}
		}
	}

	cleanup()
}

func parseSelection(selection string) ([]int, error) {
	var userIDS []int
	stringUsers := strings.Split(selection, ",")
	for _, s := range stringUsers {
		userID, err := strconv.Atoi(s)
		if !errorchecker.Check("Error parseando grupos", err) {
			userIDS = append(userIDS, userID)
		} else {
			return nil, errors.New("Error parseando usuario")
		}
	}
	fmt.Println(userIDS)
	return userIDS, nil
}

func showUsers(users []models.PublicUser) ([]int, error) {
	var usersSelected string
	if len(users) > 0 {
		for index, user := range users {
			fmt.Println(index, user.Username)
		}
		fmt.Scanf("%s", &usersSelected)
		users, err := parseSelection(usersSelected)
		if err == nil {
			return users, nil
		}
		fmt.Println("Error seleccionando usuario.")
		return nil, errors.New("Error seleccionando usuario")
	}
	fmt.Println("No se encontraron usuarios.")
	return nil, errors.New("No se encontraron usuarios.")
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

func doLogin(username string, password []byte) models.PrivateUser {
	var user models.User
	var u models.PrivateUser
	chats := map[string]models.ChatPrivateInfo{}
	passwordSlice, encrypterSlice := utils.Hash(password)
	postURL := origin + "/login"
	parameters := url.Values{"username": {username}, "pass": {utils.Encode64(passwordSlice)}}
	res, err := https.PostForm(postURL, parameters)
	if !errorchecker.Check("ERROR post", err) {
		body, err := ioutil.ReadAll(res.Body)
		if !errorchecker.Check("ERROR read body", err) {
			json.Unmarshal(body, &user)
			res.Body.Close()
			if user.Validate() {
				u = user.Parse(encrypterSlice)
				u.EncKey = encrypterSlice
				postURL := origin + "/get_state"
				parameters := url.Values{"username": {username}}
				res, err := https.PostForm(postURL, parameters)
				body, err := ioutil.ReadAll(res.Body)
				if !errorchecker.Check("ERROR read body", err) {
					body = utils.Decompress(body)
					json.Unmarshal(body, &chats)
					u.UpdateChatsInfo(chats, u.State.PrivateKey)
				}
			} else {
				fmt.Println("Error iniciando sesión, credenciales incorrectas o usuario ya logueado")
			}

		}
	}
	return u
}

func registerMenu() {
	var username string
	fmt.Println("Nombre de usuario")
	fmt.Scan(&username)
	fmt.Println("Contraseña")
	pass, _ := gopass.GetPasswd()
	user, encrypterSlice := models.RegisterUser(username, pass)
	if user.Validate() {
		currentUser = user
		currentUser.State.Chats = map[string]models.ChatInfo{}
		currentUser.EncKey = encrypterSlice
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

func cleanup() {
	postURL := origin + "/logout"
	parameters := url.Values{"username": {currentUser.Username}}
	_, err := https.PostForm(postURL, parameters)
	if !errorchecker.Check("ERROR logout", err) {
		fmt.Println(" -> Logout successful")
	}
}

func main() {
	var option string
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()
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
