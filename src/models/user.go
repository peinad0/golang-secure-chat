package models

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"project/client/src/constants"
	"project/client/src/errorchecker"
	"project/client/src/utils"
)

// User structure
type User struct {
	ID       string
	Username string
	Password string
	PubKey   string
	State    string
}

// PrivateUser structure
type PrivateUser struct {
	ID       string
	Username string
	State    State
	PubKey   *rsa.PublicKey
}

// PublicUser structure
type PublicUser struct {
	ID       string
	Username string
	PubKey   *rsa.PublicKey
}

// State user state to be retrieved by client
type State struct {
	PrivateKey *rsa.PrivateKey
	Chats      map[string]ChatInfo
}

// Validate given a user u it returns whether its attributes are valid or not
func (u *User) Validate() bool {
	if u.Username == "" {
		return false
	}
	return true
}

// Validate given a user u it returns whether its attributes are valid or not
func (u *PrivateUser) Validate() bool {
	if u.Username == "" {
		return false
	}
	return true
}

// SearchUsers given a username, it performs a http post to the server in order
// to retrieve the corresponding user or a list of users containing the given
// username in theirs
func SearchUsers(username string) []PublicUser {
	var users []User
	var publicUsers []PublicUser
	res, err := https.PostForm(constants.ServerOrigin+"/search_user", url.Values{"username": {username}})
	body, err := ioutil.ReadAll(res.Body)
	if !errorchecker.Check("ERROR in reading message", err) {
		json.Unmarshal(body, &users)
		res.Body.Close()
	}
	for _, user := range users {
		publicUsers = append(publicUsers, user.ParsePublic())
	}
	return publicUsers
}

// ParsePublic func
func (u *User) ParsePublic() PublicUser {
	var user PublicUser
	user.ID = u.ID
	user.Username = u.Username
	json.Unmarshal(utils.Decompress(utils.Decode64(u.PubKey)), &user.PubKey)
	return user
}

// GetUsers This func will return the matching user or a list of users containing
// the given username in theirs.
func GetUsers() []PublicUser {
	var username string
	fmt.Scan(&username)
	users := SearchUsers(username)
	for index, element := range users {
		fmt.Println(index, element)
	}
	return users
}

//InitializeState func
func InitializeState(privateKey *rsa.PrivateKey) []byte {
	var state State
	state.Chats = map[string]ChatInfo{}
	state.PrivateKey = privateKey
	byteState, err := json.Marshal(state)
	if !errorchecker.Check("ERROR marshaling state", err) {
		return byteState
	}
	return []byte{}
}

//AddChatToState func
func (u *PrivateUser) AddChatToState(chatinfo ChatInfo) {
	u.State.Chats[chatinfo.ChatID] = chatinfo
}

// RegisterUser function
func RegisterUser(username string, password []byte) (PrivateUser, []byte) {
	var user User
	var u PrivateUser
	passEnc, keyEnc := utils.Hash(password)
	privateKey, publicKey := utils.GetKeys()

	rawState := InitializeState(privateKey)
	rs := utils.Compress(rawState)
	encryptedState := utils.EncryptAES(rs, keyEnc)
	stateStr := utils.Encode64(encryptedState)

	publicKeyBytes, _ := json.Marshal(publicKey)
	pub := utils.Compress(publicKeyBytes)
	encodedPub := utils.Encode64(pub)
	encodedPass := utils.Encode64(passEnc)

	res, err := https.PostForm(constants.ServerOrigin+"/register", url.Values{"username": {username}, "pass": {encodedPass}, "pub": {encodedPub}, "state": {stateStr}})
	if !errorchecker.Check("ERROR post", err) {
		body, err := ioutil.ReadAll(res.Body)
		if !errorchecker.Check("Error read", err) {
			if string(body[:len(body)]) != "{error: 'user exists'}" {
				//check user values
				json.Unmarshal(body, &user)
				res.Body.Close()
				u = user.Parse(keyEnc)
			}
		}
	}
	return u, keyEnc
}

// Parse func
func (u *User) Parse(encrypter []byte) PrivateUser {
	var user PrivateUser
	user.ID = u.ID
	user.Username = u.Username
	var state State
	stateBytes := utils.Decode64(u.State)
	decrypted := utils.DecryptAES(stateBytes, encrypter)
	stateDecompressed := utils.Decompress(decrypted)
	err := json.Unmarshal(stateDecompressed, &state)
	if errorchecker.Check("ERROR unmarshal state", err) {
		user.State = State{}
	} else {
		user.State = state
		json.Unmarshal(utils.Decompress(utils.Decode64(u.PubKey)), &user.PubKey)
	}
	return user
}

//UpdateChatsInfo func
func (u *PrivateUser) UpdateChatsInfo(chats map[string]ChatPrivateInfo, key *rsa.PrivateKey) {
	var label []byte
	for _, chat := range chats {
		var info ChatInfo
		info.ChatID = chat.ChatID
		info.Token = utils.DecryptOAEP(key, utils.Decode64(chat.Token), label)
		u.State.Chats[chat.ChatID] = info
	}
}

// Print prints invoking user
func (u *User) Print() {
	fmt.Println("################### BEGIN USER ###################")
	fmt.Println(u.ID)
	fmt.Println(u.Username)
	fmt.Println(u.Password)
	fmt.Println(u.PubKey)
	fmt.Println("#################### END USER ####################")
}

// Print prints invoking user
func (s *State) Print() {
	fmt.Println("################### BEGIN State ###################")
	for _, chat := range s.Chats {
		fmt.Println(chat.ChatID)
		fmt.Println(utils.Encode64(chat.Token))
	}
	fmt.Println("#################### END State ####################")
}

// Print prints invoking user
func (u *PrivateUser) Print() {
	fmt.Println("################### BEGIN USER ###################")
	fmt.Println(u.ID)
	fmt.Println(u.Username)
	// fmt.Println(u.PrivKey)
	// fmt.Println(u.PubKey)
	fmt.Println("#################### END USER ####################")
}
