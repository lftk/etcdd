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
				break
			}
			receive <- conn
		}
	}()

	// new etcd v3 discoverer
	d, err := etcdd.NewV3(endpoints)
	if err != nil {
		return
	}
	defer d.Close()

	// list all service
	services := make(map[string]*Conn)
	svrs, err := d.Services(namespace)
	if err != nil {
		return
	}
	for key, value := range svrs {
		conn, err := connect(value, name)
		if err != nil {
			return err
		}
		services[key] = &Conn{
			name:   key,
			reader: nil,
			writer: conn,
		}
	}

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
			cherr <- errors.New("exit chat room")
		case <-ticker.C:
			if err = keepalive(); err != nil {
				cherr <- err
			}
		case ev := <-event:
			if ev.Action == "put" {
				if conn, err := connect(ev.Addr, name); err == nil {
					err = handleConn(handle, services, ev.Name, nil, conn)
				}
				if err != nil {
					cherr <- err
				}
			} else if ev.Action == "delete" {
				delete(services, ev.Name)
				handle.Delete(ev.Name)
			}
		case conn := <-receive:
			if name, err := accept(conn); err == nil {
				err = handleConn(handle, services, name, conn, nil)
			}
			if err != nil {
				cherr <- err
			}
		case err = <-cherr:
			return
		}
	}
}

func handleConn(handle Handler, services map[string]*Conn, name string, r, w net.Conn) (err error) {
	conn, ok := services[name]
	if ok {
		if r != nil {
			conn.reader = r
		}
		if w != nil {
			conn.writer = w
		}
	} else {
		conn = &Conn{
			name:   name,
			reader: r,
			writer: w,
		}
		services[name] = conn
	}

	if conn.reader != nil && conn.writer != nil {
		err = handle.New(name, conn)
	}
	return
}

func accept(conn net.Conn) (name string, err error) {
	b, err := Read(conn)
	if err != nil {
		return
	}
	name = string(b)
	return
}

func connect(addr, name string) (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", addr)
	if err != nil {
		return
	}
	err = Write(conn, []byte(name))
	return
}
