package main

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/4396/etcdd"
)

const namespace = "/services/chatroom/"

type Conn struct {
	reader net.Conn
	writer net.Conn
}

type Handler interface {
	New(name string, conn *Conn) error
	Delete(name string) error
}

// RunServer listen tcp port and register service
func RunServer(endpoints []string, name string, handle Handler) (err error) {
	// listen a local port
	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return
	}

	cherr := make(chan error, 1)
	receive := make(chan net.Conn, 16)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				cherr <- err
			}
			receive <- conn
			break
		}
	}()

	// new etcd v3 discoverer
	d, err := etcdd.NewV3(endpoints)
	if err != nil {
		return
	}
	defer d.Close()

	// register service
	addr := l.Addr().String()
	timeout := time.Second * 5
	keepalive, err := d.Register(namespace, name, addr, timeout*2)
	if err != nil {
		return
	}
	// unregister service
	defer d.Unregister(namespace, name)

	// watch service
	event, cancel, err := d.Watch(namespace)
	if err != nil {
		return
	}
	defer cancel()

	exit := make(chan os.Signal)
	signal.Notify(exit, os.Kill, os.Interrupt)

	ticker := time.NewTicker(timeout)
	defer ticker.Stop()
	for {
		select {
		case <-exit:
			cherr <- errors.New("user interrupt")
		case <-ticker.C:
			cherr <- keepalive()
		case ev := <-event:
			_ = ev
		case conn := <-receive:
			_ = conn
		case err = <-cherr:
			return
		}
	}
}
