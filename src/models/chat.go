package models

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"project/client/src/constants"
	"project/client/src/errorchecker"
	"project/client/src/utils"
	"strings"
)

// Message structure
type Message struct {
	ID      string
	Content string
	Date    string
	Sender  string
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
	ID    string
	Token string
}

//StartChat starts the chat in the server
func StartChat(sender PrivateUser, receiver PublicUser) {
	var chat Chat
	var chatInfo ChatPrivateInfo
	word := utils.RandomKey(16)
	var label []byte
	chatInfo.ID = chat.ID
	chatInfo.Token = utils.Encode64(word)
	sender.AddChatToState(chatInfo)
	receiverPubKey := receiver.PubKey
	receiverKey := utils.EncryptOAEP(receiverPubKey, word, label)
	byteSender, _ := json.Marshal(sender.State)
	stateStr := utils.Encode64(byteSender)
	res, err := http.PostForm(constants.ServerOrigin+"/new_chat", url.Values{
		"sender":      {sender.Username},
		"senderState": {stateStr},
		"receiver":    {receiver.Username},
		"receiverkey": {utils.Encode64(receiverKey)}})
	if !errorchecker.Check("ERROR post", err) {
		body, err := ioutil.ReadAll(res.Body)
		if !errorchecker.Check("ERROR read body", err) {
			json.Unmarshal(body, &chat)
			res.Body.Close()
		}
	}
	OpenChat(chat, sender)
}

//OpenChat opens the chat connection
func OpenChat(chat Chat, sender PrivateUser) {
	conn, err := net.Dial("tcp", "localhost:1337") // llamamos al servidor
	if err != nil {
		fmt.Println("ERROR", err)
	}

	defer conn.Close() // es importante cerrar la conexión al finalizar

	fmt.Println()
	fmt.Println(chat.Name+"(", conn.RemoteAddr(), " - ", conn.LocalAddr(), ")")
	fmt.Println()

	names := strings.Split(chat.Name, " ")
	var name string // name of the other person

	if names[0] == sender.Username {
		name = names[2]
	} else {
		name = names[0]
	}

	keyscan := bufio.NewScanner(os.Stdin) // scanner para la entrada estándar (teclado)
	netscan := bufio.NewScanner(conn)     // scanner para la conexión (datos desde el servidor)

	if len(chat.Messages) > 0 {
		for _, msg := range chat.Messages {
			if msg.Sender == sender.ID {
				fmt.Println("Yo: " + msg.Content)
			} else {
				fmt.Println(name + ": " + msg.Content)
			}
		}
	}

	// Send chat info to the server
	chatInfo, _ := json.Marshal(chat)
	fmt.Fprintln(conn, utils.Encode64(chatInfo))

	// Send user info to the server
	userInfo, _ := json.Marshal(sender)
	fmt.Fprintln(conn, utils.Encode64(userInfo))

	go func() {
		for netscan.Scan() { // escaneamos la conexión
			fmt.Println(name + ": " + netscan.Text()) // mostramos mensaje desde el servidor
		}
	}()

	for keyscan.Scan() { // escaneamos la entrada
		text := keyscan.Text()
		if text == "/exit" {
			break
		}
		fmt.Fprintln(conn, keyscan.Text()) // enviamos la entrada al servidor
	}

}

//GetChats get the list of chats the use has
func GetChats(user PrivateUser) ([]Chat, error) {
	var chats []Chat
	res, err := http.PostForm(constants.ServerOrigin+"/get_chats", url.Values{"userid": {user.ID}})
	body, err := ioutil.ReadAll(res.Body)
	if !errorchecker.Check("ERROR in reading message", err) {
		json.Unmarshal(body, &chats)
		res.Body.Close()
	}
	return chats, err
}
