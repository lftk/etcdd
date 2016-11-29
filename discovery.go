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
var TimeoutUnit = time.Second * 3

// A CancelFunc tells an operation to abandon its work.
type CancelFunc context.CancelFunc

// KeepAlive keep service online
func KeepAlive(endpoints []string, namespace, name, addr string) (cancel CancelFunc, err error) {
	api, err := newKeysAPI(endpoints)
	if err == nil {
		cancel, err = KeepAliveWithKeysAPI(api, namespace, name, addr)
	}
	return
}

// KeepAliveWithKeysAPI use client.KeysAPI
func KeepAliveWithKeysAPI(api client.KeysAPI, namespace, name, addr string) (CancelFunc, error) {
	key := filepath.Join(namespace, name)
	_, err := api.Create(context.Background(), key, addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(2 * TimeoutUnit)
		options := &client.SetOptions{
			PrevExist:        client.PrevExist,
			TTL:              3 * TimeoutUnit,
			Refresh:          true,
			NoValueOnSuccess: true,
		}

		defer func() {
			api.Delete(context.Background(), key, nil)
			ticker.Stop()
		}()

		for {
			_, err := api.Set(ctx, key, "", options)
			if err != nil {
				break
			}
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}
	}()
	return CancelFunc(cancel), nil
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
	api, err := newKeysAPI(endpoints)
	if err == nil {
		event, cancel, err = WatchWithKeysAPI(api, namespace)
	}
	return
}

// WatchWithKeysAPI use etcd.Client.KeysAPI
func WatchWithKeysAPI(api client.KeysAPI, namespace string) (<-chan *Event, CancelFunc, error) {
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
	return event, CancelFunc(cancel), nil
}

// Services list all service
func Services(endpoints []string, namespace string) (svrs map[string]string, err error) {
	api, err := newKeysAPI(endpoints)
	if err == nil {
		svrs, err = ServicesWithKeysAPI(api, namespace)
	}
	return
}

// ServicesWithKeysAPI use etcd.client.KeysAPI
func ServicesWithKeysAPI(api client.KeysAPI, namespace string) (svrs map[string]string, err error) {
	resp, err := api.Get(context.Background(), namespace, &client.GetOptions{
		Recursive: true,
	})
	if err != nil {
		return
	}

	var name string
	svrs = make(map[string]string)
	for _, node := range resp.Node.Nodes {
		name, err = filepath.Rel(namespace, node.Key)
		if err != nil {
			break
		}
		svrs[name] = node.Value
	}
	return
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
