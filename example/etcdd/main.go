package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/4396/etcdd"
	"github.com/4396/etcdd/api"
)

var (
	svrname = flag.String("name", "hello", "service name")
	svraddr = flag.String("addr", "127.0.0.1:1234", "service address")
	etcdver = flag.Int("ver", 3, "service address")
)

func main() {
	flag.Parse()

	namespace := "/services"
	endpoints := []string{
		"http://127.0.0.1:2379",
	}

	var err error
	var d api.Discoverer
	if *etcdver == 2 {
		d, err = etcdd.NewV2(endpoints)
	} else if *etcdver == 3 {
		d, err = etcdd.NewV3(endpoints)
	} else {
		log.Fatal("invalid etcd version")
	}

	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		d.Close()
	}()

	svrs, err := d.Services(namespace)
	if err != nil {
		log.Fatal(err)
	}
	for name, addr := range svrs {
		fmt.Println(name, addr)
	}

	timeout := time.Second * 5
	keepalive, err := d.Register(namespace, *svrname, *svraddr, timeout*2)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		d.Unregister(namespace, *svrname)
	}()

	event, cancel, err := d.Watch(namespace)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		cancel()
	}()

	ticker := time.NewTicker(timeout)
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
			err := keepalive()
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
