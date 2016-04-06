package main

import (
    "fmt"
    "log"

    "golang.org/x/net/websocket"
)

var origin = "http://go-secure-chat.herokuapp.com"
var url = "ws://go-secure-chat.herokuapp.com/chat"

func main() {
    ws, err := websocket.Dial(url, "", origin)
    if err != nil {
        log.Fatal(err)
    }

    message := []byte("Chat funcionando!")
    _, err = ws.Write(message)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Send: %s\n", message)

    var msg = make([]byte, 512)
    _, err = ws.Read(msg)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Receive: %s\n", msg)
}