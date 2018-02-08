package models

import (
	"errors"
	"fmt"
	"log"
	"project/server/src/constants"
	"project/server/src/errorchecker"
	"project/server/src/utils"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// User structure
type User struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Username string        `bson:"name"`
	Password string        `bson:"password"`
	Salt     string        `bson:"salt"`
	PubKey   string        `bson:"pubkey"`
	State    string        `bson:"state"`
}

// PublicUser structure
type PublicUser struct {
	ID       bson.ObjectId
	Username string
	PubKey   string
}

// State user state to be retrieved by client
type State struct {
	PrivateKey string
	Chats      []ChatPrivateInfo
	Contacts   []PublicUser
}

//SetSalt func
func (u *User) SetSalt(salt string) {
	u.Salt = salt
}

//GetSalt func
func (u *User) GetSalt() string {
	return u.Salt
}

//GetUsernames func
func GetUsernames(ids []string) map[string]string {
	usernames := make(map[string]string)

	for _, id := range ids {
		user := GetByID(bson.ObjectIdHex(id))
		usernames[id] = user.Username
	}
	return usernames
}

//GetByID func
func GetByID(id bson.ObjectId) User {
	var user User
	session, err := mgo.Dial(constants.URI)
	errorchecker.Check("ERROR dialing", err)
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	usersCollection := session.DB(constants.AuthDatabase).C("user")
	err = usersCollection.FindId(id).One(&user)
	if !errorchecker.Check("ERROR searching user", err) {
		return user
	}
	return User{}
}

// Login given a username and password, it tries to return its info from DB
func Login(username string, password []byte) User {
	user := SearchUser(username)
	if user.Validate() {
		salt := utils.Decode64(user.Salt)
		hashedPasswd, err := utils.ScryptHash(password, salt)
		if err == nil {
			if user.Password == utils.Encode64(hashedPasswd) {
				return user
			}
		}
	}
	return User{}
}

func extractUsers(tokens []ChatToken) []bson.ObjectId {
	var userIDS []bson.ObjectId
	for _, token := range tokens {
		user := SearchUser(token.Username)
		userIDS = append(userIDS, user.ID)
	}
	return userIDS
}

// RegisterUser registered
func RegisterUser(username, password, pubKey, state string) (User, error) {
	var user User
	var returnError error
	users := SearchUsers(username)
	if len(users) == 1 {
		returnError = errors.New("Username taken")
	} else {
		decodedPassword := utils.Decode64(password)
		hashedPassword, salt := utils.HashWithRandomSalt(decodedPassword)

		user.Username = username
		user.Password = utils.Encode64(hashedPassword)
		user.Salt = utils.Encode64(salt)
		user.PubKey = pubKey
		user.State = state
		user = user.Save()
	}
	return user, returnError
}

// Save func
func (u *User) Save() User {
	var user User
	session, err := mgo.Dial(constants.URI)
	errorchecker.Check("ERROR dialing", err)
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB(constants.AuthDatabase).C("user")
	err = c.Insert(&u)
	errorchecker.Check("ERROR inserting user", err)
	user = SearchUser(u.Username)
	return user
}

// UpdateState func
func (u *User) UpdateState() {
	session, err := mgo.Dial(constants.URI)
	errorchecker.Check("ERROR dialing", err)
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB(constants.AuthDatabase).C("user")
	update := bson.M{"$set": bson.M{"state": u.State}}
	err = c.UpdateId(u.ID, update)
	errorchecker.Check("ERROR inserting user", err)
}

// SearchUsers returns the list of User objects containing the given username as part of theirs
func SearchUsers(username string) []PublicUser {
	var users []User
	var publicUsers []PublicUser
	session, err := mgo.Dial(constants.URI)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	usersCollection := session.DB(constants.AuthDatabase).C("user")
	err = usersCollection.Find(bson.M{"name": bson.RegEx{username, "i"}}).All(&users)
	errorchecker.Check("ERROR searching users", err)
	for _, user := range users {
		publicUsers = append(publicUsers, user.GetPublicUser())
	}
	return publicUsers
}

// SearchUser returns the User object given a user with username
func SearchUser(username string) User {
	var user User
	session, err := mgo.Dial(constants.URI)
	errorchecker.Check("ERROR dialing", err)
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	usersCollection := session.DB(constants.AuthDatabase).C("user")
	err = usersCollection.Find(bson.M{"name": username}).One(&user)
	if !errorchecker.Check("ERROR searching user", err) {
		return user
	}
	return User{}
}

// SearchUser returns the User object given a user with id
func SearchUserById(id bson.ObjectId) User {
	var user User
	session, err := mgo.Dial(constants.URI)
	errorchecker.Check("ERROR dialing", err)
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	usersCollection := session.DB(constants.AuthDatabase).C("user")
	err = usersCollection.Find(bson.M{"_id": id}).One(&user)
	if !errorchecker.Check("ERROR searching user", err) {
		return user
	}
	return User{}
}

// GetPublicUser function
func (u *User) GetPublicUser() PublicUser {
	var user PublicUser
	user.ID = u.ID
	user.Username = u.Username
	user.PubKey = u.PubKey
	return user
}

//AddChat adds the chat token to the user profile
func (u *User) AddChat(chatid bson.ObjectId, token string) {
	session, err := mgo.Dial(constants.URI)
	errorchecker.Check("ERROR dialing", err)
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB(constants.AuthDatabase).C("user")
	colQuerier := bson.M{"name": u.Username}
	change := bson.M{"$push": bson.M{"chats": bson.M{"id": chatid, "token": token}}}
	err = c.Update(colQuerier, change)
	errorchecker.Check("ERROR inserting chat", err)
}

//GetUsersByID func
func GetUsersByID(users []bson.ObjectId) []PublicUser {
	var returnUsers []PublicUser
	var user User
	for _, ID := range users {
		user = GetByID(ID)
		returnUsers = append(returnUsers, user.GetPublicUser())
	}
	return returnUsers
}

// Validate given a user u it returns whether its attributes are valid or not
func (u *User) Validate() bool {
	if u.Username == "" {
		return false
	}
	return true
}

// Print prints invoking user
func (u *User) Print() {
	fmt.Println("################### USER #####################")
	fmt.Println(u.ID)
	fmt.Println(u.Username)
	fmt.Println(u.Password)
	fmt.Println(len(u.PubKey))
	fmt.Println(u.State)
	fmt.Println(u.Salt)
	fmt.Println("################# END USER ###################")
}
