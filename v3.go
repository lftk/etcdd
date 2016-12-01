package discovery

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/coreos/etcd/clientv3"
)

// A V3 object wrap Client
type V3 struct {
	cli *clientv3.Client
}

// NewV3 returns v3 discoverer
func NewV3(endpoints []string) (*V3, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		return nil, err
	}
	v3 := &V3{
		cli: cli,
	}
	return v3, nil
}

// Register service
func (v *V3) Register(namespace, name, addr string) (KeepAliveFunc, error) {
	key := filepath.Join(namespace, name)
	resp, err := v.cli.Grant(context.TODO(), 5)
	if err != nil {
		return nil, err
	}
	_, err = v.cli.KeepAlive(context.TODO(), resp.ID)
	if err != nil {
		return nil, err
	}
	_, err = v.cli.Put(context.TODO(), key, addr, clientv3.WithLease(resp.ID))
	if err != nil {
		return nil, err
	}
	return v.keepAlive(resp.ID), nil
}

func (v *V3) keepAlive(id clientv3.LeaseID) KeepAliveFunc {
	return func(timeout time.Duration) error {
		_, err := v.cli.KeepAliveOnce(context.TODO(), id)
		return err
	}
}

// Unregister service
func (v *V3) Unregister(namespace, name string) error {
	key := filepath.Join(namespace, name)
	_, err := v.cli.Delete(context.TODO(), key)
	return err
}

// Watch services, returns event and cancel func
func (v *V3) Watch(namespace string) (<-chan *Event, CancelFunc, error) {
	event := make(chan *Event, 1024)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer close(event)

		wc := v.cli.Watch(ctx, namespace, clientv3.WithPrefix())
		for resp := range wc {
			for _, ev := range resp.Events {
				fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}
		}
	}()
	return event, CancelFunc(cancel), nil
}

// Services list all service
func (v *V3) Services(namespace string) (map[string]string, error) {
	return nil, nil
}
