package discovery

import (
	"context"
	"path/filepath"
	"time"

	"github.com/coreos/etcd/client"
)

// KeepAliveFunc need to execute on timer
type KeepAliveFunc func(time.Duration) error

// A CancelFunc tells an operation to abandon its work.
type CancelFunc context.CancelFunc

// Event desc service status change
// Action: create set update expire delete
type Event struct {
	Action string
	Name   string
	Addr   string
}

// A V2 object contains Client and KeysAPI
type V2 struct {
	cli client.Client
	api client.KeysAPI
}

// NewV2 returns v2 discoverer
func NewV2(endpoints []string) (*V2, error) {
	cli, err := client.New(client.Config{
		Endpoints:               endpoints,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: client.DefaultRequestTimeout,
	})
	if err != nil {
		return nil, err
	}
	api := client.NewKeysAPI(cli)
	v2 := &V2{
		cli: cli,
		api: api,
	}
	return v2, nil
}

// Register service
func (v *V2) Register(namespace, name, addr string) (KeepAliveFunc, error) {
	key := filepath.Join(namespace, name)
	_, err := v.api.Create(context.TODO(), key, addr)
	if err != nil {
		return nil, err
	}
	return v.keepAlive(key), nil
}

func (v *V2) keepAlive(key string) KeepAliveFunc {
	return func(timeout time.Duration) error {
		_, err := v.api.Set(context.Background(), key, "", &client.SetOptions{
			PrevExist:        client.PrevExist,
			TTL:              2 * timeout,
			Refresh:          true,
			NoValueOnSuccess: true,
		})
		return err
	}
}

// Unregister service
func (v *V2) Unregister(namespace, name string) error {
	key := filepath.Join(namespace, name)
	_, err := v.api.Delete(context.TODO(), key, nil)
	return err
}

// Watch services, returns event and cancel func
func (v *V2) Watch(namespace string) (<-chan *Event, CancelFunc, error) {
	event := make(chan *Event, 1024)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer close(event)

		watcher := v.api.Watcher(namespace, &client.WatcherOptions{
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
func (v *V2) Services(namespace string) (map[string]string, error) {
	resp, err := v.api.Get(context.TODO(), namespace, &client.GetOptions{
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
