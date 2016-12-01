package main

import (
	"net"
)

type server struct {
	name     string
	addr     string
	listener net.Listener
}

func newServer(name, addr string) (s *server) {
	s = &server{
		name: name,
		addr: addr,
	}
	return s
}

func (s *server) Run(handle func(conn net.Conn) error) (err error) {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return
	}
	for {
		var conn net.Conn
		conn, err = l.Accept()
		if err != nil {
			break
		}
		err = handle(conn)
		if err != nil {
			break
		}
	}
	return
}
