package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
)

func main() {
	endpoints := []string{
		"127.0.0.1:2379",
	}

	name := randomName()
	fmt.Println(name)

	log.Fatal(RunServer(endpoints, name, nil))
}

func randomName() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(b)
}
