package models

import (
	"bufio"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"project/client/src/constants"
	"project/client/src/errorchecker"
	"project/client/src/utils"
	"strconv"
	"strings"
	"time"
)

// Message structure
type Message struct {
	Content []byte
	Type    MessageType
	Date    string
	Sender  string
}

//MessageType struct
type MessageType struct {
	Type string
	Name string
}

//Chat structure
type Chat struct {
	ID         string
	Components []string
	Admin      string
	Messages   []Message
	Name       string
	Type       string
}

// ChatPrivateInfo struct
type ChatPrivateInfo struct {
	Username string
	ChatID   string
	Token    string
}

// ChatToken struct
type ChatToken struct {
	Username string
	Token    string
}

// ChatInfo struct
type ChatInfo struct {
	ChatID string
	Token  []byte
}

var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var https = &http.Client{Transport: tr}

//StartChat starts the chat in the server
func StartChat(sender PrivateUser, receivers []PublicUser) {
	var chat Chat
	var tokens []ChatToken
	var token ChatToken
	var chatInfo ChatInfo
	word := utils.RandomKey(32)
	var label []byte
	var name string

	for _, receiver := range receivers {
		receiverKey := utils.EncryptOAEP(receiver.PubKey, word, label)
		token.Token = utils.Encode64(receiverKey)
		token.Username = receiver.Username
		tokens = append(tokens, token)
	}
	marshaledTokens, _ := json.Marshal(tokens)
	strTokens := utils.Encode64(marshaledTokens)
	var chatType string
	if len(receivers) == 1 {
		chatType = "individual"
		name = sender.Username + "-" + receivers[0].Username
	} else {
		chatType = "group"
		fmt.Println("Nombre del grupo: ")
		reader := bufio.NewReader(os.Stdin)
		name, _ = reader.ReadString('\n')
		if len(name) == 0 {
			name = "(Undefined)"
		}
	}
	fmt.Println(name)
	res, err := https.PostForm(constants.ServerOrigin+"/new_chat", url.Values{
		"sender": {sender.Username},
		"type":   {chatType},
		"name":   {name},
		"tokens": {strTokens}})

	if !errorchecker.Check("ERROR post", err) {
		body, err := ioutil.ReadAll(res.Body)
		if !errorchecker.Check("ERROR read body", err) {
			json.Unmarshal(body, &chat)
			chatInfo.Token = word
			chatInfo.ChatID = chat.ID
			sender.AddChatToState(chatInfo)
			sender.UpdateState()
			res.Body.Close()
		}
	}
	OpenChat(chat, sender)
}

func downloadFile(username string, message Message, key []byte) error {
	descifrado := utils.DecryptAES(message.Content, key)
	filename := message.Type.Name
	wd, err := os.Getwd()
	errorchecker.Check("Error getting working directory", err)
	err = ioutil.WriteFile(wd+"/Downloads/"+username+"-"+filename, descifrado, 0644)
	return err
}

//Print func
func (m *Message) Print() {
	fmt.Println("********** MESSAGE **********")
	fmt.Println(m.Sender)
	fmt.Println(m.Date)
	fmt.Println(m.Type.Type)
	fmt.Println("*****************************")

}

func handleMessage(username string, msg Message, key []byte) {
	//msg.Print()
	var sender string
	switch msg.Type.Type {
	case "text":
		descifrado := utils.DecryptAES(msg.Content, key)
		if msg.Sender == username {
			sender = "Yo"
		} else {
			sender = msg.Sender
		}
		fmt.Printf("[%s] %s: %s\n", msg.Date, sender, descifrado)
		break
	case "file":
		err := downloadFile(username, msg, key)
		if !errorchecker.Check("Error descargando archivo", err) {
			fmt.Printf("[%s] %s: Archivo recibido: %s\n", msg.Date, msg.Sender, msg.Type.Name)
		}
		break
	}
}

//GetAdminChats func
func GetAdminChats(admin string) []Chat {
	var chats []Chat
	postURL := constants.ServerOrigin + "/get_admin_chats"
	parameters := url.Values{"username": {admin}}
	res, err := https.PostForm(postURL, parameters)
	body, err := ioutil.ReadAll(res.Body)
	if !errorchecker.Check("ERROR read body", err) {
		json.Unmarshal(body, &chats)
	}
	return chats
}

//OpenChat opens the chat connection
func OpenChat(chat Chat, sender PrivateUser) {
	conn, err := tls.Dial("tcp", "localhost:1337", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		fmt.Println("ERROR", err)
	}

	defer conn.Close() // es importante cerrar la conexión al finalizar

	fmt.Println()
	fmt.Println(chat.Name+"(", conn.RemoteAddr(), " - ", conn.LocalAddr(), ")")
	fmt.Println()
	key := sender.State.Chats[chat.ID].Token

	keyscan := bufio.NewScanner(os.Stdin) // scanner para la entrada estándar (teclado)
	netscan := bufio.NewScanner(conn)     // scanner para la conexión (datos desde el servidor)

	if len(chat.Messages) > 0 {
		for _, msg := range chat.Messages {
			handleMessage(sender.Username, msg, key)
		}
	}

	// Send chat info to the server
	chatInfo, _ := json.Marshal(chat)
	fmt.Fprintln(conn, utils.Encode64(chatInfo))

	// Send user info to the server
	userInfo, _ := json.Marshal(sender)
	fmt.Fprintln(conn, utils.Encode64(userInfo))

	go func() {
		var msg Message
		for netscan.Scan() { // escaneamos la conexión
			data := netscan.Text()
			receivedData := utils.Decode64(data)
			json.Unmarshal(receivedData, &msg)
			handleMessage(sender.Username, msg, key)
		}
	}()

	var canSendMessage bool
	var exit bool
	var t MessageType
	var message Message
	message.Sender = sender.Username

	for keyscan.Scan() { // escaneamos la entrada
		text := keyscan.Text()
		canSendMessage = true
		exit = false
		message.Date = time.Now().Format("2006-01-02 15:04:05")
		switch text {
		case "/exit":
			// Exit conversation
			exit = true
			canSendMessage = false
			break
		case "/send_file":
			// Send file entering path
			fmt.Println("Path del archivo a enviar:")
			keyscan.Scan()
			filename := keyscan.Text()
			file, err := ioutil.ReadFile(filename)
			if !errorchecker.Check("ERROR ReadFile path", err) {
				t.Type = "file"
				t.Name = message.Sender + "-" + time.Now().Format("20060102150405") + filepath.Ext(filename)
				message.Type = t
				fmt.Println("sexo")
				message.Content = utils.EncryptAES(file, key)
			} else {
				canSendMessage = false
			}
			break
		default:
			// Normal message
			bytesText := []byte(text)
			t.Type = "text"
			t.Name = ""
			message.Type = t
			message.Content = utils.EncryptAES(bytesText, key)
			break
		}
		if canSendMessage {
			data, err := json.Marshal(message)
			sendData := utils.Encode64(data)
			if !errorchecker.Check("ERROR Marshaling", err) {
				fmt.Fprintln(conn, sendData) // enviamos la entrada al servidor
			}
		}
		if exit {
			break
		}
	}
}

//GetChatUsers func
func GetChatUsers(components []string) []PublicUser {
	type tmpUser struct {
		ID       string
		Username string
		PubKey   string
	}
	var usernames []tmpUser
	var user PublicUser
	var users []PublicUser
	var pubKey *rsa.PublicKey
	postURL := constants.ServerOrigin + "/get_chat_users"
	bytesComponents, _ := json.Marshal(components)
	parameters := url.Values{"users": {utils.Encode64(bytesComponents)}}
	res, err := https.PostForm(postURL, parameters)
	body, err := ioutil.ReadAll(res.Body)
	if !errorchecker.Check("ERROR read body", err) {
		json.Unmarshal(body, &usernames)
	}

	for _, u := range usernames {
		user.ID = u.ID
		user.Username = u.Username
		pubBytes := utils.Decode64(u.PubKey)
		data := utils.Decompress(pubBytes)
		json.Unmarshal(data, &pubKey)
		user.PubKey = pubKey
		users = append(users, user)
	}

	return users
}

//AdminMenu func
func AdminMenu() {
	fmt.Println("¿Qué acción deseas realizar?")
	fmt.Println("1. Agregar a la conversación")
	fmt.Println("2. Expulsar de la conversación")
	fmt.Println("q. Salir")
}
func parseSelection(selection string) ([]int, error) {
	var userIDS []int
	stringUsers := strings.Split(selection, ",")
	for _, s := range stringUsers {
		userID, err := strconv.Atoi(s)
		if !errorchecker.Check("Error parseando usuarios", err) {
			userIDS = append(userIDS, userID)
		} else {
			return nil, errors.New("Error parseando usuario")
		}
	}
	fmt.Println(userIDS)
	return userIDS, nil
}
func showUsers(users []PublicUser) ([]int, error) {
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

//DeleteUser func
func (c *Chat) DeleteUser(users []PublicUser, selected []int) Chat {
	var chat Chat
	var deleteUsers []PublicUser
	for _, i := range selected {
		deleteUsers = append(deleteUsers, users[i])
	}
	postURL := constants.ServerOrigin + "/delete_chat_users"
	bytesUsers, _ := json.Marshal(deleteUsers)
	chatBytes, _ := json.Marshal(c)
	chatStr := utils.Encode64(chatBytes)
	parameters := url.Values{
		"users": {utils.Encode64(bytesUsers)},
		"chat":  {chatStr}}
	res, err := https.PostForm(postURL, parameters)
	body, err := ioutil.ReadAll(res.Body)
	if !errorchecker.Check("Error read", err) {
		json.Unmarshal(body, &chat)
		res.Body.Close()
	}
	return chat
}

//UpdateKey func
func (c *Chat) UpdateKey(admin *PrivateUser) {
	var tokens []ChatToken
	var token ChatToken
	var chat Chat
	var chatInfo ChatInfo
	word := utils.RandomKey(32)
	var label []byte

	chatInfo.ChatID = c.ID
	chatInfo.Token = word
	admin.State.Chats[c.ID] = chatInfo
	admin.UpdateState()

	receivers := GetChatUsers(c.Components)

	for _, receiver := range receivers {
		receiverKey := utils.EncryptOAEP(receiver.PubKey, word, label)
		token.Token = utils.Encode64(receiverKey)
		token.Username = receiver.Username
		tokens = append(tokens, token)
	}
	marshaledTokens, _ := json.Marshal(tokens)
	strTokens := utils.Encode64(marshaledTokens)

	res, err := https.PostForm(constants.ServerOrigin+"/update_chat_key", url.Values{
		"chat":   {c.ID},
		"tokens": {strTokens}})

	if !errorchecker.Check("ERROR post", err) {
		body, err := ioutil.ReadAll(res.Body)
		if !errorchecker.Check("ERROR read body", err) {
			json.Unmarshal(body, &chat)
			res.Body.Close()
		}
	}
	fmt.Println(chat.Components)
}

//AdministrarChat func
func AdministrarChat(admin *PrivateUser, chat Chat) {
	fmt.Println("Panel de administración de", admin.Username)
	var c string
	AdminMenu()
	fmt.Scanf("%s", &c)
	switch c {
	case "1":
		break
	case "2":
		users := GetChatUsers(chat.Components)
		selection, _ := showUsers(users)
		updatedChat := chat.DeleteUser(users, selection)
		updatedChat.UpdateKey(admin)
		break
	}
}

//GetChats get the list of chats the use has
func GetChats(user PrivateUser) ([]Chat, error) {
	chatsInfo := map[string]ChatPrivateInfo{}
	postURL := constants.ServerOrigin + "/get_state"
	parameters := url.Values{"username": {user.Username}}
	res, err := https.PostForm(postURL, parameters)
	body, err := ioutil.ReadAll(res.Body)
	if !errorchecker.Check("ERROR read body", err) {
		body = utils.Decompress(body)
		json.Unmarshal(body, &chatsInfo)
		user.UpdateChatsInfo(chatsInfo, user.State.PrivateKey)
	}

	var chats []Chat
	res, err = https.PostForm(constants.ServerOrigin+"/get_chats", url.Values{"userid": {user.ID}})
	body, err = ioutil.ReadAll(res.Body)
	if !errorchecker.Check("ERROR in reading message", err) {
		json.Unmarshal(body, &chats)
		res.Body.Close()
	}
	return chats, err
}
