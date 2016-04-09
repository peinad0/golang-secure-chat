package src

import (
	"bytes"
	"net/http"
)

// Login sends the credentials to the server to log into de system
func Login(username string, pasword string) bool {
	credentials := []byte(`{"username": username, "password": password}`)
	_, err := http.Post(ServerOrigin+"/login", "application/json", bytes.NewBuffer(credentials))
	if err != nil {
		return false
	}
	return true
}
