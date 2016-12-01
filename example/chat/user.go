package main

import (
	"net"
)

type user struct {
	name    string
	friends map[string]net.Conn
}

func (u *user) Name() string {
	return u.name
}

func (u *user) Hello() string {
	return "hello"
}

func (u *user) Bye() string {
	return "bye"
}
