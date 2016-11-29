package main

import (
	"fmt"
	"log"

	etcdd "github.com/4396/etcd-discovery"
)

func main() {
	endpoints := []string{
		"http://127.0.0.1:2379",
	}
	_, err := etcdd.KeepAlive(endpoints, "/services", "hello", "127.0.0.1:1234")
	if err != nil {
		log.Fatal(err)
	}

	event, cancel, err := etcdd.Watch(endpoints, "/services")
	if err != nil {
		log.Fatal(err)
	}
	for {
		ev, ok := <-event
		if !ok {
			break
		}
		fmt.Println(ev.Action, ev.Name, ev.Addr)
	}
	cancel()
}
