package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
)

func main() {
	endpoints := []string{
		"127.0.0.1:2379",
	}

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	name := hex.EncodeToString(b)

	user := NewUser(name)
	go func() {
		r := bufio.NewReader(os.Stdin)
		for {
			line, _, err := r.ReadLine()
			if err == nil {
				err = user.Send(line)
			}
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	err = RunServer(endpoints, name, user)
	if err != nil {
		log.Fatal(err)
	}
}
