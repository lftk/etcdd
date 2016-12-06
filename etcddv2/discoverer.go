package etcddv2

import (
	"context"
	"path/filepath"
	"time"

	"github.com/4396/etcdd/api"
	"github.com/coreos/etcd/client"
)

// A Discoverer object contains Client and KeysAPI
type Discoverer struct {
	cli client.Client
	api client.KeysAPI
}

// New returns v2 discoverer
func New(endpoints []string) (*Discoverer, error) {
	return NewWithAuth(endpoints, "", "")
}

// NewWithAuth need username and password
func NewWithAuth(endpoints []string, username, password string) (*Discoverer, error) {
	cli, err := client.New(client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: client.DefaultRequestTimeout,
		Username:                username,
		Password:                password,
	})
	if err != nil {
		return nil, err
	}
	api := client.NewKeysAPI(cli)
	d := &Discoverer{
		cli: cli,
		api: api,
	}
	return d, nil
}

// Close discoverer
func (d *Discoverer) Close() error {
	return nil
}

// Register service
func (d *Discoverer) Register(namespace, name, addr string, timeout time.Duration) (api.KeepAliveFunc, error) {
	key := filepath.Join(namespace, name)
	_, err := d.api.Set(context.TODO(), key, addr, &client.SetOptions{
		PrevExist:        client.PrevNoExist,
		TTL:              timeout,
		Refresh:          false,
		NoValueOnSuccess: true,
	})
	if err != nil {
		return nil, err
	}
	return d.keepAlive(key, timeout), nil
}

func (d *Discoverer) keepAlive(key string, timeout time.Duration) api.KeepAliveFunc {
	return func() error {
		_, err := d.api.Set(context.TODO(), key, "", &client.SetOptions{
			PrevExist:        client.PrevExist,
			TTL:              timeout,
			Refresh:          true,
			NoValueOnSuccess: true,
		})
		return err
	}
}

// Unregister service
func (d *Discoverer) Unregister(namespace, name string) error {
	key := filepath.Join(namespace, name)
	_, err := d.api.Delete(context.TODO(), key, nil)
	return err
}

// Watch services, returns event and cancel func
func (d *Discoverer) Watch(namespace string) (<-chan *api.Event, api.CancelFunc, error) {
	event := make(chan *api.Event, 1024)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer close(event)

		watcher := d.api.Watcher(namespace, &client.WatcherOptions{
			Recursive: true,
		})

		for {
			resp, err := watcher.Next(ctx)
			if err != nil {
				if err == context.Canceled {
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
			event <- &api.Event{
				Action: resp.Action,
				Name:   name,
				Addr:   node.Value,
			}
		}
	}()
	return event, api.CancelFunc(cancel), nil
}

// Services list all service
func (d *Discoverer) Services(namespace string) (map[string]string, error) {
	resp, err := d.api.Get(context.TODO(), namespace, &client.GetOptions{
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

// Version returns etcd client version
func (d *Discoverer) Version() int {
	return 2
}
