package discovery

import (
	"context"
	"path/filepath"
	"time"

	"github.com/coreos/etcd/client"
)

// KeepAliveFunc need to execute on timer
type KeepAliveFunc func(time.Duration) error

// Register service
func Register(endpoints []string, namespace, name, addr string) (KeepAliveFunc, error) {
	api, err := newKeysAPI(endpoints)
	if err != nil {
		return nil, err
	}
	return register(api, namespace, name, addr)
}

func register(api client.KeysAPI, namespace, name, addr string) (KeepAliveFunc, error) {
	key := filepath.Join(namespace, name)
	_, err := api.Create(context.Background(), key, addr)
	if err != nil {
		return nil, err
	}
	return func(timeout time.Duration) error {
		_, err = api.Set(context.Background(), key, "", &client.SetOptions{
			PrevExist:        client.PrevExist,
			TTL:              2 * timeout,
			Refresh:          true,
			NoValueOnSuccess: true,
		})
		return err
	}, nil
}

// Unregister service
func Unregister(endpoints []string, namespace, name string) error {
	api, err := newKeysAPI(endpoints)
	if err != nil {
		return err
	}
	return unregister(api, namespace, name)
}

func unregister(api client.KeysAPI, namespace, name string) error {
	key := filepath.Join(namespace, name)
	_, err := api.Delete(context.Background(), key, nil)
	return err
}

// Event desc service status change
// Action: create set update expire delete
type Event struct {
	Action string
	Name   string
	Addr   string
}

// A CancelFunc tells an operation to abandon its work.
type CancelFunc context.CancelFunc

// Watch a service, returns event and cancel func
func Watch(endpoints []string, namespace string) (<-chan *Event, CancelFunc, error) {
	api, err := newKeysAPI(endpoints)
	if err != nil {
		return nil, nil, err
	}
	return watch(api, namespace)
}

func watch(api client.KeysAPI, namespace string) (<-chan *Event, CancelFunc, error) {
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
				if cerr := ctx.Err(); cerr != nil {
					break
				}
				time.Sleep(client.DefaultRequestTimeout)
				continue
			}

			node := resp.Node
			if resp.Action == "expire" || resp.Action == "delete" {
				node = resp.PrevNode
			}

			name, _ := filepath.Rel(namespace, node.Key)
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
func Services(endpoints []string, namespace string) (map[string]string, error) {
	api, err := newKeysAPI(endpoints)
	if err != nil {
		return nil, err
	}
	return services(api, namespace)
}

func services(api client.KeysAPI, namespace string) (map[string]string, error) {
	resp, err := api.Get(context.Background(), namespace, &client.GetOptions{
		Recursive: true,
	})
	if err != nil {
		return nil, err
	}

	var name string
	svrs := make(map[string]string)
	for _, node := range resp.Node.Nodes {
		name, err = filepath.Rel(namespace, node.Key)
		if err != nil {
			break
		}
		svrs[name] = node.Value
	}
	return svrs, nil
}

func newKeysAPI(endpoints []string) (client.KeysAPI, error) {
	cli, err := client.New(client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: client.DefaultRequestTimeout,
	})
	if err != nil {
		return nil, err
	}
	return client.NewKeysAPI(cli), nil
}
