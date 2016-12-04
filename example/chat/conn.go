package main

import (
	"net"
)

type Conn struct {
	name   string
	reader net.Conn
	writer net.Conn
}

func (c *Conn) Send(b []byte) (err error) {
	err = Write(c.writer, b)
	return
}

func (c *Conn) Receive(handle func([]byte) error) (err error) {
	var b []byte
	for err == nil {
		b, err = Read(c.reader)
		if err == nil {
			err = handle(b)
		}
	}
	return
}
