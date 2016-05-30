package utils

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"project/client/src/errorchecker"

	"golang.org/x/net/websocket"
)

// Compress función para comprimir
func Compress(data []byte) []byte {
	var b bytes.Buffer      // b contendrá los datos comprimidos (tamaño variable)
	w := zlib.NewWriter(&b) // escritor que comprime sobre b
	w.Write(data)           // escribimos los datos
	w.Close()               // cerramos el escritor (buffering)
	return b.Bytes()        // devolvemos los datos comprimidos
}

// Decompress función para descomprimir
func Decompress(data []byte) []byte {
	var b bytes.Buffer // b contendrá los datos descomprimidos

	r, err := zlib.NewReader(bytes.NewReader(data)) // lector descomprime al leer

	errorchecker.Check("ERROR comprimiendo", err) // comprobamos el error
	io.Copy(&b, r)                                // copiamos del descompresor (r) al buffer (b)
	r.Close()                                     // cerramos el lector (buffering)
	return b.Bytes()                              // devolvemos los datos descomprimidos
}

// Encode64 función para codificar de []bytes a string (Base64)
func Encode64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data) // sólo utiliza caracteres "imprimibles"
}

// Decode64 función para decodificar de string a []bytes (Base64)
func Decode64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)   // recupera el formato original
	errorchecker.Check("ERROR decodificando", err) // comprobamos el error
	return b                                       // devolvemos los datos originales
}

func send(ws *websocket.Conn, m string) {

	message := []byte(m)
	_, err := ws.Write(message)
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
