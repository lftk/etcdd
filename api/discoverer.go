package api

import (
	"context"
	"time"
)

// KeepAliveFunc need to execute on timer
type KeepAliveFunc func() error

// A CancelFunc tells an operation to abandon its work.
type CancelFunc context.CancelFunc

// Event desc service status change
// Action:
// 	v2 - create set update expire delete
//	v3 - put delete
type Event struct {
	Action string
	Name   string
	Addr   string
}

// Discoverer interface
type Discoverer interface {
	Register(namespace, name, addr string, timeout time.Duration) (KeepAliveFunc, error)
	Unregister(namespace, name string) error
	Watch(namespace string) (<-chan *Event, CancelFunc, error)
	Services(namespace string) (map[string]string, error)
	Close() error
	Version() int
}
