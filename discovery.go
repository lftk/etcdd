package discovery

import (
	"context"
	"path/filepath"
	"time"

	"github.com/coreos/etcd/client"
)

// TimeoutUnit desc a time unit
// HeaderTimeoutPerRequest = 1 * TimeoutUnit
// Update service ticker = 2 * TimeoutUnit
// SetOptions.TTL = 3 * TimeoutUnit
var TimeoutUnit = time.Second * 10

// A CancelFunc tells an operation to abandon its work.
type CancelFunc context.CancelFunc

// KeepAlive keep service online
func KeepAlive(endpoints []string, namespace, name, addr string) (cancel CancelFunc, err error) {
	if api, err := newKeysAPI(endpoints); err == nil {
		cancel = KeepAliveWithKeysAPI(api, namespace, name, addr)
	}
	return
}

// KeepAliveWithKeysAPI use client.KeysAPI
func KeepAliveWithKeysAPI(api client.KeysAPI, namespace, name, addr string) CancelFunc {
	key := filepath.Join(namespace, name)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer api.Delete(context.Background(), key, nil)

		ticker := time.NewTicker(2 * TimeoutUnit)
		options := &client.SetOptions{
			TTL: 3 * TimeoutUnit,
		}
		for {
			_, err := api.Set(ctx, key, addr, options)
			if err != nil {
				break
			}
			select {
			case <-ctx.Done():
				break
			case <-ticker.C:
			}
		}
	}()
	return CancelFunc(cancel)
}

// Event desc service status change
// Action: set update expire delete
type Event struct {
	Action string
	Name   string
	Addr   string
}

// Watch a service, returns event and cancel func
func Watch(endpoints []string, namespace string) (event <-chan *Event, cancel CancelFunc, err error) {
	if api, err := newKeysAPI(endpoints); err == nil {
		event, cancel = WatchWithKeysAPI(api, namespace)
	}
	return
}

// WatchWithKeysAPI use etcd.Client.KeysAPI
func WatchWithKeysAPI(api client.KeysAPI, namespace string) (<-chan *Event, CancelFunc) {
	event := make(chan *Event, 1024)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer close(event)

		watcher := api.Watcher(namespace, &client.WatcherOptions{
			Recursive: true,
		})

		for {
			resp, err := watcher.Next(ctx)
			if err != nil {
				break
			}

			node := resp.Node
			if resp.Action == "expire" || resp.Action == "delete" {
				node = resp.PrevNode
			}

			name, err := filepath.Rel(namespace, node.Key)
			if err != nil {
				break
			}
			event <- &Event{
				Action: resp.Action,
				Name:   name,
				Addr:   node.Value,
			}
		}
	}()
	return event, CancelFunc(cancel)
}

// new etcd.client.KeysAPI
func newKeysAPI(endpoints []string) (api client.KeysAPI, err error) {
	cli, err := client.New(client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: TimeoutUnit,
	})
	if err == nil {
		api = client.NewKeysAPI(cli)
	}
	return
}
