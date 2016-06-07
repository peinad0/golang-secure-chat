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
	"project/client/src/constants"
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
func StartChat(sender PrivateUser, receivers []PublicUser, encrypterKey []byte) {
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
			byteSender, _ := json.Marshal(sender.State)
			compressed := utils.Compress(byteSender)
			encrypted := utils.EncryptAES(compressed, encrypterKey)
			stateStr := utils.Encode64(encrypted)
			res.Body.Close()
			https.PostForm(constants.ServerOrigin+"/update_state", url.Values{
				"username": {sender.Username},
				"state":    {stateStr}})
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
			descifrado := utils.DecryptAES(utils.Decode64(msg.Content), key)
			fmt.Printf("%s\n", descifrado)
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

			fmt.Printf("%s\n", descifrado) // mostramos mensaje desde el servidor
		}
	}()

	for keyscan.Scan() { // escaneamos la entrada
		text := sender.Username + ": " + keyscan.Text()
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
