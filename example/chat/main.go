package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
)

func main() {
	endpoints := []string{
		"127.0.0.1:2379",
	}

	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	name := hex.EncodeToString(b)
	fmt.Println(name)

	log.Fatal(RunServer(endpoints, name, callback))
}

func callback(name string, send net.Conn, receive net.Conn) error {
	if send != nil {
		fmt.Println(name, send.RemoteAddr().String())
	}
	if receive != nil {
		fmt.Println(name, receive.RemoteAddr().String())
	}
	return nil
}
