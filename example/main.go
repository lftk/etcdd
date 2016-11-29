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

	cancel, err := etcdd.KeepAlive(endpoints, "/services", "hello", "127.0.0.1:1234")
	if err != nil {
		log.Fatal(err)
	}
	defer cancel()

	event, _, err := etcdd.Watch(endpoints, "/services")
	if err != nil {
		log.Fatal(err)
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Kill, os.Interrupt)

	select {
	case <-exit:
		break
	case ev, ok := <-event:
		if ok {
			fmt.Println(ev.Action, ev.Name, ev.Addr)
		} else {
			break
		}
	}
	fmt.Println("---------bye----------")
}
