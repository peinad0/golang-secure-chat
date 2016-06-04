package models

import (
	"crypto/rsa"
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
	PubKey   string
	PrivKey  string
}

// PrivateUser structure
type PrivateUser struct {
	ID       string
	Username string
	PubKey   *rsa.PublicKey
	PrivKey  *rsa.PrivateKey
}

// PublicUser structure
type PublicUser struct {
	ID       string
	Username string
	PubKey   *rsa.PublicKey
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
	res, err := http.PostForm(constants.ServerOrigin+"/search_user", url.Values{"username": {username}})
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

// RegisterUser function
func RegisterUser(username string, password []byte) PrivateUser {
	var user User
	var u PrivateUser
	passEnc, keyEnc := utils.Hash(password)
	rawPriv, _ := utils.GetKeys()
	priv, pub := utils.CipherKeys(rawPriv, keyEnc)
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
				u = user.Parse(keyEnc)
			}
		}
	}
	return u
}

// Parse func
func (u *User) Parse(encrypter []byte) PrivateUser {
	var user PrivateUser
	user.ID = u.ID
	user.Username = u.Username
	json.Unmarshal(utils.Decompress(utils.DecryptAES(utils.Decode64(u.PrivKey), encrypter)), &user.PrivKey)
	json.Unmarshal(utils.Decompress(utils.Decode64(u.PubKey)), &user.PubKey)
	return user
}

// Print prints invoking user
func (u *User) Print() {
	fmt.Println("################### BEGIN USER ###################")
	fmt.Println(u.ID)
	fmt.Println(u.Username)
	fmt.Println(u.Password)
	fmt.Println(u.PrivKey)
	fmt.Println(u.PubKey)
	fmt.Println("#################### END USER ####################")
}

// Print prints invoking user
func (u *PrivateUser) Print() {
	fmt.Println("################### BEGIN USER ###################")
	fmt.Println(u.ID)
	fmt.Println(u.Username)
	fmt.Println(u.PrivKey)
	fmt.Println(u.PubKey)
	fmt.Println("#################### END USER ####################")
}
