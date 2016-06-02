package models

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
	"project/client/src/chatloadenv"
	"project/client/src/errorchecker"
	"project/client/src/utils"
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

//StartChat starts the chat in the server
func StartChat(sender, receiver User) {
	var chat Chat
	word := utils.RandomKey(16)

	receiverPubKey := utils.Decode64(receiver.PubKey)
	receiverKey := utils.EncryptAES(word, receiverPubKey[:32])
	senderPubKey := utils.Decode64(sender.PubKey)
	senderKey := utils.EncryptAES(word, senderPubKey[:32])

	res, err := http.PostForm(chatloadenv.ServerOrigin+"/new_chat", url.Values{
		"sender":      {sender.Username},
		"senderkey":   {base64.StdEncoding.EncodeToString(senderKey)},
		"receiver":    {receiver.Username},
		"receiverkey": {base64.StdEncoding.EncodeToString(receiverKey)}})

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
func OpenChat(chat Chat, sender User) {
	conn, err := net.Dial("tcp", "localhost:1337") // llamamos al servidor
	if err != nil {
		fmt.Println("ERROR", err)
	}

	defer conn.Close() // es importante cerrar la conexión al finalizar

	fmt.Println()
	fmt.Println(chat.Name+"(", conn.RemoteAddr(), ")")
	fmt.Println()

	keyscan := bufio.NewScanner(os.Stdin) // scanner para la entrada estándar (teclado)
	netscan := bufio.NewScanner(conn)     // scanner para la conexión (datos desde el servidor)

	if len(chat.Messages) > 0 {
		for _, msg := range chat.Messages {
			if msg.Sender == sender.ID {
				fmt.Println("Yo: " + msg.Content)
			} else {
				fmt.Println(sender.Username + ": " + msg.Content)
			}
		}
	}

	// Send chat info to the server
	chatInfo, _ := json.Marshal(chat)
	fmt.Fprintln(conn, base64.StdEncoding.EncodeToString(chatInfo))

	// Send user info to the server
	userInfo, _ := json.Marshal(sender)
	fmt.Fprintln(conn, base64.StdEncoding.EncodeToString(userInfo))

	for keyscan.Scan() { // escaneamos la entrada
		text := keyscan.Text()
		if text == "/exit" {
			break
		}
		fmt.Fprintln(conn, keyscan.Text())         // enviamos la entrada al servidor
		netscan.Scan()                             // escaneamos la conexión
		fmt.Println("Response: " + netscan.Text()) // mostramos mensaje desde el servidor
	}
}

//GetChats get the list of chats the use has
func GetChats(user User) ([]Chat, error) {
	var chats []Chat
	res, err := http.PostForm(chatloadenv.ServerOrigin+"/get_chats", url.Values{"userid": {user.ID}})
	body, err := ioutil.ReadAll(res.Body)
	if !errorchecker.Check("ERROR in reading message", err) {
		json.Unmarshal(body, &chats)
		res.Body.Close()
	}
	return chats, err
}
