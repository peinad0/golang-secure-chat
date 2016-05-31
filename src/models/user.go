package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"project/client/src/chatloadenv"
	"project/client/src/errorchecker"
	"project/client/src/utils"
)

// User structure
type User struct {
	ID       string
	Username string
	Password string
	Salt     string
	PubKey   string
	PrivKey  string
}

// Login given a user, it tries to return its info from DB
func (u *User) Login() {
	fmt.Println("a")
}

// Validate given a user u it returns whether its attributes are valid or not
func (u *User) Validate() bool {
	if u.Username == "" {
		return false
	}
	return true
}

// SearchUser given a username, it performs a http post to the server in order
// to retrieve the corresponding user or a list of users containing the given
// username in theirs
func SearchUser(username string) ([]User, error) {
	var users []User
	res, err := http.PostForm(chatloadenv.ServerOrigin+"/search_user", url.Values{"username": {username}})
	body, err := ioutil.ReadAll(res.Body)
	if !errorchecker.Check("ERROR in reading message", err) {
		json.Unmarshal(body, &users)
		res.Body.Close()
	}
	return users, err
}

// GetUsers This func will return the matching user or a list of users containing
// the given username in theirs.
func GetUsers() []User {
	var username string
	fmt.Scan(&username)
	users, _ := SearchUser(username)

	// DEBUGGING
	for index, element := range users {
		fmt.Println(index, element)
	}
	return users
}

// RegisterUser function
func RegisterUser(username string, password []byte) User {
	var user User
	passEnc, keyEnc := utils.Hash(password)
	pub, priv := utils.GetKeys(keyEnc)
	encodedPass := utils.Encode64(passEnc)
	encodedPub := utils.Encode64(pub)
	encodedPriv := utils.Encode64(priv)
	res, err := http.PostForm(chatloadenv.ServerOrigin+"/register", url.Values{"username": {username}, "pass": {encodedPass}, "pub": {encodedPub}, "priv": {encodedPriv}})
	if !errorchecker.Check("ERROR post", err) {
		body, err := ioutil.ReadAll(res.Body)
		if !errorchecker.Check("Error read", err) {
			if string(body[:len(body)]) != "{error: 'user exists'}" {
				json.Unmarshal(body, &user)
				res.Body.Close()
			}
		}
		fmt.Print(body)
	}
	return user
}

// Print prints invoking user
func (u *User) Print() {
	fmt.Println("################### USER #####################")
	fmt.Println(u.ID)
	fmt.Println(u.Username)
	fmt.Println(u.Password)
	fmt.Println(u.PrivKey)
	fmt.Println(u.PubKey)
	fmt.Println(u.Salt)
	fmt.Println("################# END USER ###################")
}
