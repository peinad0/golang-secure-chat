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
	Username string
	ChatID   string
	Token    string
}

// ChatInfo struct
type ChatInfo struct {
	ChatID string
	Token  []byte
}

//StartChat starts the chat in the server
func StartChat(sender PrivateUser, receiver PublicUser, encrypterKey []byte) {
	var chat Chat
	var chatInfo ChatInfo
	word := utils.RandomKey(32)
	var label []byte
	receiverPubKey := receiver.PubKey
	receiverKey := utils.EncryptOAEP(receiverPubKey, word, label)
	fmt.Println(receiverKey)
	res, err := http.PostForm(constants.ServerOrigin+"/new_chat", url.Values{
		"sender":      {sender.Username},
		"receiver":    {receiver.Username},
		"receiverkey": {utils.Encode64(receiverKey)}})
	if !errorchecker.Check("ERROR post", err) {
		body, err := ioutil.ReadAll(res.Body)
		if !errorchecker.Check("ERROR read body", err) {
			json.Unmarshal(body, &chat)
			chatInfo.Token = word
			chatInfo.ChatID = chat.ID
			sender.AddChatToState(chatInfo)
			byteSender, _ := json.Marshal(sender.State)
			compressed := utils.Compress(byteSender)
			encrypted := utils.EncryptAES(compressed, encrypterKey)
			stateStr := utils.Encode64(encrypted)
			res.Body.Close()
			http.PostForm(constants.ServerOrigin+"/update_state", url.Values{
				"username": {sender.Username},
				"state":    {stateStr}})
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
	key := sender.State.Chats[chat.ID].Token

	names := strings.Split(chat.Name, " ")
	var name string // name of the other person
	token := sender.State.Chats[chat.ID].Token
	fmt.Println(utils.Encode64(token))
	if names[0] == sender.Username {
		name = names[2]
	} else {
		name = names[0]
	}

	keyscan := bufio.NewScanner(os.Stdin) // scanner para la entrada estándar (teclado)
	netscan := bufio.NewScanner(conn)     // scanner para la conexión (datos desde el servidor)

	if len(chat.Messages) > 0 {
		for _, msg := range chat.Messages {
			descifrado := utils.DecryptAES(utils.Decode64(msg.Content), key)
			if msg.Sender == sender.ID {
				fmt.Printf("Yo: %s\n", descifrado)
			} else {
				fmt.Printf("%s: %s\n", name, descifrado)
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
			text := netscan.Text()
			descifrado := utils.DecryptAES(utils.Decode64(text), key)

			fmt.Printf("%s: %s\n", name, descifrado) // mostramos mensaje desde el servidor
		}
	}()

	for keyscan.Scan() { // escaneamos la entrada
		text := keyscan.Bytes()
		if keyscan.Text() == "/exit" {
			break
		}

		cifrado := utils.EncryptAES([]byte(text), key)

		fmt.Fprintln(conn, utils.Encode64(cifrado)) // enviamos la entrada al servidor
	}

}

//GetChats get the list of chats the use has
func GetChats(user PrivateUser) ([]Chat, error) {
	chatsInfo := map[string]ChatPrivateInfo{}
	postURL := constants.ServerOrigin + "/get_state"
	parameters := url.Values{"username": {user.Username}}
	res, err := http.PostForm(postURL, parameters)
	body, err := ioutil.ReadAll(res.Body)
	if !errorchecker.Check("ERROR read body", err) {
		body = utils.Decompress(body)
		json.Unmarshal(body, &chatsInfo)
		user.UpdateChatsInfo(chatsInfo, user.State.PrivateKey)
	}

	var chats []Chat
	res, err = http.PostForm(constants.ServerOrigin+"/get_chats", url.Values{"userid": {user.ID}})
	body, err = ioutil.ReadAll(res.Body)
	if !errorchecker.Check("ERROR in reading message", err) {
		json.Unmarshal(body, &chats)
		res.Body.Close()
	}
	return chats, err
}
