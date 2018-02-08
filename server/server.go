package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"project/client/src/errorchecker"
	"project/server/src/constants"
	"project/server/src/models"
	"project/server/src/utils"
	"strconv"

	"gopkg.in/mgo.v2/bson"

	"golang.org/x/net/websocket"
)

// start docs
// godoc -http=:6060
var connectedUsers map[string]models.User

func chatHandler(ws *websocket.Conn) {
	msg := make([]byte, 512)
	n, err := ws.Read(msg)
	if err != nil {
		log.Fatal(err)
	}
	message := string(msg[:n])
	log.Printf("Receive: %s\n", message)

	response := ""

	_, err = ws.Write([]byte(response))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Send: %s\n", response)
}

func checkLogin(username string, passwd []byte) models.User {
	fmt.Println("Usuario <" + username + "> intenta loguearse.")
	user := models.SearchUser(username)

	if user.Validate() {
		salt := utils.Decode64(user.GetSalt())
		hashedPasswd, err := utils.ScryptHash(passwd, salt)
		if err == nil {
			if user.Password == utils.Encode64(hashedPasswd) {
				fmt.Println("Login correcto.")
				addConnectedUser(user)
				return user
			}
		}
	}
	fmt.Println("Login usuario <" + username + "> rechazado.")
	return models.User{}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Check user chats and return them
	username := r.FormValue("username")
	pass := utils.Decode64(r.FormValue("pass"))
	user := checkLogin(username, pass)
	res, _ := json.Marshal(user)
	w.Write(res)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	fmt.Println("Usuario <" + username + "> intenta registrarse.")
	password := r.FormValue("pass")
	pubKey := r.FormValue("pub")
	state := r.FormValue("state")
	user, err := models.RegisterUser(username, password, pubKey, state)
	if err != nil {
		fmt.Println("Usuario <" + username + "> registro rechazado.")
		w.Write([]byte("{error: 'user exists'}"))
	} else {
		fmt.Println("Usuario <" + username + "> registrado correctamente.")
		addConnectedUser(user)
		res, _ := json.Marshal(user)
		w.Write(res)
	}
}

func searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	users := models.SearchUsers(username)
	res, _ := json.Marshal(users)
	w.Write(res)
}

func newChatHandler(w http.ResponseWriter, r *http.Request) {
	var receivers []models.User
	var tokens []models.ChatToken
	senderUsername := r.FormValue("sender")
	chatType := r.FormValue("type")
	name := r.FormValue("name")
	receiversTokens := utils.Decode64(r.FormValue("tokens"))
	json.Unmarshal(receiversTokens, &tokens)
	sender := models.SearchUser(senderUsername)
	for _, token := range tokens {
		user := models.SearchUser(token.Username)
		receivers = append(receivers, user)
	}
	chatid := models.CreateChat(sender, receivers, name, chatType)
	models.SaveChatInfo(tokens, chatid)
	var chat models.Chat
	chat = models.GetChat(chatid.Hex())
	res, _ := json.Marshal(chat)
	w.Write(res)
}

func getChatsHandler(w http.ResponseWriter, r *http.Request) {
	userid := r.FormValue("userid")
	chats := models.GetChats(userid)
	res, _ := json.Marshal(chats)
	w.Write(res)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	fmt.Println("Usuario <" + username + "> intenta hacer logout.")
	removeConnectedUser(username)
	fmt.Println("Logout correcto")
	w.Write([]byte("Logout"))
}

func printConnectedUsers() {
	index := 1
	fmt.Println("Hay " + strconv.Itoa(len(connectedUsers)) + " usuarios conectados:")
	for key := range connectedUsers {
		fmt.Println(strconv.Itoa(index) + ": " + key)
		index++
	}
}

func addConnectedUser(user models.User) {
	connectedUsers[user.Username] = user
	printConnectedUsers()
}

func removeConnectedUser(username string) {
	delete(connectedUsers, username)
	printConnectedUsers()
}

func getStateHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	fmt.Println("Usuario <" + username + "> intenta recuperar su estado.")
	chatsInfo := models.RecuperarEstado(username)
	byteChats, err := json.Marshal(chatsInfo)
	if errorchecker.Check("ERROR Marshal state", err) {
		fmt.Println("Usuario <" + username + "> error recuperando su estado.")
	}
	byteChats = utils.Compress(byteChats)
	fmt.Println("Usuario <" + username + "> ha recuperado su estado.")
	w.Write(byteChats)
}

func updateStateHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	state := r.FormValue("state")

	fmt.Println("Usuario <" + username + "> intenta actualizar su estado.")
	user := models.SearchUser(username)
	user.State = state
	user.UpdateState()
	w.Write([]byte("OK"))
}

func getAdminChatsHandler(w http.ResponseWriter, r *http.Request) {
	var chats []models.Chat
	username := r.FormValue("username")
	chats = models.GetAdminChats(username)
	res, _ := json.Marshal(chats)
	w.Write(res)
}

func getChatUsersHandler(w http.ResponseWriter, r *http.Request) {
	var usernames []bson.ObjectId
	var users []models.PublicUser

	data := r.FormValue("users")
	bytes := utils.Decode64(data)
	json.Unmarshal(bytes, &usernames)
	users = models.GetUsersByID(usernames)
	res, _ := json.Marshal(users)
	w.Write(res)
}

func deleteChatUsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []models.PublicUser
	var chat models.Chat
	usersStr := r.FormValue("users")
	chatStr := r.FormValue("chat")
	dataUser := utils.Decode64(usersStr)
	dataChat := utils.Decode64(chatStr)
	json.Unmarshal(dataUser, &users)
	json.Unmarshal(dataChat, &chat)
	chat.DeleteUsers(users)
	chatBytes, _ := json.Marshal(chat)
	w.Write(chatBytes)
}

func addChatUsersHandler(w http.ResponseWriter, r *http.Request) {
	var tokens []models.ChatToken
	var chat models.Chat
	tokensStr := r.FormValue("tokens")
	chatStr := r.FormValue("chat")
	dataTokens := utils.Decode64(tokensStr)
	dataChat := utils.Decode64(chatStr)
	json.Unmarshal(dataTokens, &tokens)
	json.Unmarshal(dataChat, &chat)

	chat.AddUsers(tokens)

	chatBytes, _ := json.Marshal(chat)
	w.Write(chatBytes)
}

func updateChatKeyHandler(w http.ResponseWriter, r *http.Request) {
	var receivers []models.User
	var tokens []models.ChatToken
	chatid := r.FormValue("chat")
	receiversTokens := utils.Decode64(r.FormValue("tokens"))
	json.Unmarshal(receiversTokens, &tokens)
	for _, token := range tokens {
		user := models.SearchUser(token.Username)
		receivers = append(receivers, user)
	}
	models.SaveChatInfo(tokens, bson.ObjectIdHex(chatid))
	w.Write([]byte("UPDATED"))
}

func main() {
	connectedUsers = make(map[string]models.User)
	http.Handle("/chat", websocket.Handler(chatHandler))
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/search_user", searchUsersHandler)
	http.HandleFunc("/new_chat", newChatHandler)
	http.HandleFunc("/get_chats", getChatsHandler)
	http.HandleFunc("/update_chat_key", updateChatKeyHandler)
	http.HandleFunc("/get_admin_chats", getAdminChatsHandler)
	http.HandleFunc("/get_state", getStateHandler)
	http.HandleFunc("/get_chat_users", getChatUsersHandler)
	http.HandleFunc("/add_chat_users", addChatUsersHandler)
	http.HandleFunc("/update_state", updateStateHandler)
	http.HandleFunc("/delete_chat_users", deleteChatUsersHandler)

	go models.OpenChat(connectedUsers)
	err := http.ListenAndServeTLS(constants.Port, "cert.pem", "key.pem", nil)
	if err != nil {
		fmt.Println("ListenAndServe error")
	}
}
