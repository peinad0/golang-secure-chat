package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	Salt     string
	PubKey   string
	PrivKey  string
}

// PublicUser structure
type PublicUser struct {
	Username string
	PubKey   string
}

// Validate given a user u it returns whether its attributes are valid or not
func (u *User) Validate() bool {
	if u.Username == "" {
		return false
	}
	return true
}

// SearchUsers given a username, it performs a http post to the server in order
// to retrieve the corresponding user or a list of users containing the given
// username in theirs
func SearchUsers(username string) ([]PublicUser, error) {
	var users []PublicUser
	res, err := http.PostForm(constants.ServerOrigin+"/search_user", url.Values{"username": {username}})
	body, err := ioutil.ReadAll(res.Body)
	if !errorchecker.Check("ERROR in reading message", err) {
		json.Unmarshal(body, &users)
		res.Body.Close()
	}

	return users, err
}

// GetUsers This func will return the matching user or a list of users containing
// the given username in theirs.
func GetUsers() []PublicUser {
	var username string
	fmt.Scan(&username)
	users, _ := SearchUsers(username)

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
	res, err := http.PostForm(constants.ServerOrigin+"/register", url.Values{"username": {username}, "pass": {encodedPass}, "pub": {encodedPub}, "priv": {encodedPriv}})
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
	fmt.Println("################### BEGING USER ###################")
	fmt.Println(u.ID)
	fmt.Println(u.Username)
	fmt.Println(u.Password)
	fmt.Println(u.PrivKey)
	fmt.Println(u.PubKey)
	fmt.Println(u.Salt)
	fmt.Println("#################### END USER #####################")
}
