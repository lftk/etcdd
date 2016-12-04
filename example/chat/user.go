package main

import "fmt"
import "errors"

type User struct {
	name    string
	friends map[string]*Conn
}

func NewUser(name string) *User {
	return &User{
		name:    name,
		friends: make(map[string]*Conn),
	}
}

func (u *User) Name() string {
	return u.name
}

func (u *User) New(name string, conn *Conn) (err error) {
	fmt.Println("--", name, "join chat room")
	u.friends[name] = conn
	err = conn.Send([]byte("hello"))
	if err != nil {
		return
	}
	go func() {
		conn.Receive(func(b []byte) error {
			fmt.Println(name, ":", string(b))
			return nil
		})
	}()
	return
}

func (u *User) Delete(name string) (err error) {
	fmt.Println("--", name, "exit chat room")
	delete(u.friends, name)
	return
}

func (u *User) Send(b []byte) (err error) {
	for name, friend := range u.friends {
		err = friend.Send(b)
		if err != nil {
			err = errors.New(name + err.Error())
			break
		}
	}
	return
}
