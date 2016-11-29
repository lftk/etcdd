package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	etcdd "github.com/4396/etcd-discovery"
)

func main() {
	endpoints := []string{
		"http://127.0.0.1:2379",
	}

	svrs, err := etcdd.Services(endpoints, "/services")
	if err != nil {
		log.Fatal(err)
	}
	for name, addr := range svrs {
		fmt.Println(name, addr)
	}

	err = etcdd.Register(endpoints, "/services", "hello", "127.0.0.1:1234")
	if err != nil {
		log.Fatal(err)
	}
	defer etcdd.Unregister(endpoints, "/services", "hello")

	event, _, err := etcdd.Watch(endpoints, "/services")
	if err != nil {
		log.Fatal(err)
	}

	exit := make(chan os.Signal)
	signal.Notify(exit, os.Kill, os.Interrupt)

	for {
		select {
		case <-exit:
			goto EXIT
		case ev, ok := <-event:
			if !ok {
				goto EXIT
			}
			fmt.Println(ev.Action, ev.Name, ev.Addr)
		}
	}

EXIT:
	fmt.Println("---------bye----------")
}
