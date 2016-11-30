package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	etcdd "github.com/4396/etcd-discovery"
)

func main() {
	svrname := "hello"
	svraddr := "127.0.0.1:1234"
	namespace := "/services"
	endpoints := []string{
		"http://127.0.0.1:2379",
	}

	v2etcdd, err := etcdd.NewV2(endpoints)
	if err != nil {
		log.Fatal(err)
	}

	svrs, err := v2etcdd.Services(namespace)
	if err != nil {
		log.Fatal(err)
	}
	for name, addr := range svrs {
		fmt.Println(name, addr)
	}

	keepalive, err := v2etcdd.Register(namespace, svrname, svraddr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		v2etcdd.Unregister(namespace, svrname)
	}()

	event, cancel, err := v2etcdd.Watch(namespace)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		cancel()
	}()

	ticker := time.NewTicker(time.Second * 5)
	defer func() {
		ticker.Stop()
	}()

	exit := make(chan os.Signal)
	signal.Notify(exit, os.Kill, os.Interrupt)

	for {
		select {
		case <-exit:
			goto EXIT
		case <-ticker.C:
			err := keepalive(time.Second * 5)
			if err != nil {
				log.Fatal(err)
			}
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
