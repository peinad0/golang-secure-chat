/*

Este programa demuestra una arquitectura cliente servidor sencilla.
El cliente envía líneas desde la entrada estandar y el servidor le devuelve un reconomiento de llegada (acknowledge).
El servidor es concurrente, siendo capaz de manejar múltiples clientes simultáneamente.
Las entradas se procesan mediante un scanner (bufio).

ejemplos de uso:

go run cnx.go srv

go run cnx.go cli

*/

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

// función para comprobar errores (ahorra escritura)
func chk(e error) {
	if e != nil {
		panic(e)
	}
}

func getCredentials() (user, pass string) {
	fmt.Printf("User: ")

	_, err := fmt.Fscanf(os.Stdin, "%s", &user)

	if err != nil {
		fmt.Printf(err.Error())
	}

	fmt.Printf("Password: ")

	_, err = fmt.Fscanf(os.Stdin, "%s", &pass)

	if err != nil {
		fmt.Printf(err.Error())
	}

	return
}

func main() {
	fmt.Println("Iniciando cliente...")
	user, pass := getCredentials()
	if login(user, pass) {
		client()
	}
}

func client() {
	conn, err := net.Dial("tcp", "localhost:1337") // llamamos al servidor
	chk(err)
	defer conn.Close() // es importante cerrar la conexión al finalizar

	fmt.Println("conectado a ", conn.RemoteAddr())

	keyscan := bufio.NewScanner(os.Stdin) // scanner para la entrada estándar (teclado)
	netscan := bufio.NewScanner(conn)     // scanner para la conexión (datos desde el servidor)

	for keyscan.Scan() { // escaneamos la entrada
		fmt.Fprintln(conn, keyscan.Text())         // enviamos la entrada al servidor
		netscan.Scan()                             // escaneamos la conexión
		fmt.Println("servidor: " + netscan.Text()) // mostramos mensaje desde el servidor
	}
}

func login(user string, pass string) bool {
	if user == "admin" && pass == "admin" {
		return true
	}
	return false
}
