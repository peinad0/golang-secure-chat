package models

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"project/client/src/constants"
	"project/client/src/errorchecker"
	"project/client/src/utils"
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
	} else {
		chatType = "group"
	}
	res, err := https.PostForm(constants.ServerOrigin+"/new_chat", url.Values{
		"sender": {sender.Username},
		"type":   {chatType},
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

//GetChatUsernames func
func GetChatUsernames(components []string) map[string]string {
	var usernames map[string]string
	postURL := constants.ServerOrigin + "/get_chat_names"
	bytesComponents, _ := json.Marshal(components)
	parameters := url.Values{"users": {utils.Encode64(bytesComponents)}}
	res, err := https.PostForm(postURL, parameters)
	body, err := ioutil.ReadAll(res.Body)
	if !errorchecker.Check("ERROR read body", err) {
		json.Unmarshal(body, &usernames)
	}
	return usernames
}

func downloadFile(message Message, key []byte) error {
	descifrado := utils.DecryptAES(message.Content, key)
	filename := message.Type.Name
	wd, err := os.Getwd()
	errorchecker.Check("Error getting working directory", err)
	err = ioutil.WriteFile(wd+"/Downloads/"+filename, descifrado, 0644)
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

func handleMessage(msg Message, key []byte) {
	//msg.Print()
	switch msg.Type.Type {
	case "text":
		descifrado := utils.DecryptAES(msg.Content, key)
		fmt.Printf("[%s] %s: %s\n", msg.Date, msg.Sender, descifrado)
		break
	case "file":
		err := downloadFile(msg, key)
		if !errorchecker.Check("Error descargando archivo", err) {
			fmt.Printf("[%s] %s: Archivo recibido: %s\n", msg.Date, msg.Sender, msg.Type.Name)
		}
		break
	}
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

	//names := GetChatUsernames(chat.Components)

	keyscan := bufio.NewScanner(os.Stdin) // scanner para la entrada estándar (teclado)
	netscan := bufio.NewScanner(conn)     // scanner para la conexión (datos desde el servidor)

	if len(chat.Messages) > 0 {
		for _, msg := range chat.Messages {
			handleMessage(msg, key)
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
			handleMessage(msg, key)
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
		switch text {
		case "/exit":
			// Exit conversation
			exit = true
			break
		case "/send_file":
			// Send file entering path
			fmt.Println("Path del archivo a enviar:")
			keyscan.Scan()
			filename := keyscan.Text()
			file, err := ioutil.ReadFile(filename)
			if !errorchecker.Check("ERROR ReadFile path", err) {
				message.Date = time.Now().String()
				t.Type = "file"
				t.Name = message.Sender + "-" + time.Now().String() + filepath.Ext(filename)
				message.Type = t
				message.Content = utils.EncryptAES(file, key)
			} else {
				canSendMessage = false
			}
			break
		default:
			// Normal message
			bytesText := []byte(text)
			message.Date = time.Now().String()
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
