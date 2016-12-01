package main

import (
	"net"

	"time"

	"github.com/4396/etcdd"
)

const namespace = "/services/chatroom/"

type Callback func(name string, send net.Conn, receive net.Conn) error

// RunServer listen tcp port and register service
func RunServer(endpoints []string, name string, callback Callback) (err error) {
	// listen tcp port
	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return
	}
	defer l.Close()

	receive := make(chan net.Conn, 16)
	cherr := make(chan error, 1)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				cherr <- err
				break
			}
			receive <- conn
		}
	}()

	// register, keepalive and watch service
	d, err := etcdd.NewV3(endpoints)
	if err != nil {
		return
	}

	svrs, err := d.Services(namespace)
	if err != nil {
		return
	}
	for name, addr := range svrs {
		var conn net.Conn
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			return
		}
		err = callback(name, conn, nil)
		if err != nil {
			return
		}
	}

	timeout := time.Second * 5
	keepalive, err := d.Register(namespace, name, l.Addr().String(), timeout*2)
	if err != nil {
		return
	}
	defer d.Unregister(namespace, name)

	event, cancel, err := d.Watch(namespace)
	if err != nil {
		return
	}
	defer cancel()

	ticker := time.NewTicker(timeout)
	defer ticker.Stop()
	for {
		select {
		case err = <-cherr:
			return
		case <-ticker.C:
			err = keepalive()
			if err != nil {
				return
			}
		case ev := <-event:
			if ev.Action == "put" {
				var conn net.Conn
				conn, err = net.Dial("tcp", ev.Addr)
				if err != nil {
					return
				}
				err = callback(ev.Name, conn, nil)
				if err != nil {
					return
				}
			} else if ev.Action == "delete" {

			}
		case conn := <-receive:
			err = callback("", nil, conn)
			if err != nil {
				return
			}
		}
	}
}
