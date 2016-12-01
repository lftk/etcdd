package etcddv3

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/4396/etcdd/api"
	"github.com/coreos/etcd/clientv3"
)

// A Discoverer wrap Client
type Discoverer struct {
	cli *clientv3.Client
}

// New returns v3 discoverer
func New(endpoints []string) (*Discoverer, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		return nil, err
	}
	d := &Discoverer{
		cli: cli,
	}
	return d, nil
}

// Close discoverer
func (d *Discoverer) Close() error {
	return d.cli.Close()
}

// Register service
func (d *Discoverer) Register(namespace, name, addr string, timeout time.Duration) (api.KeepAliveFunc, error) {
	ttl := int64(timeout.Seconds())
	resp, err := d.cli.Grant(context.TODO(), ttl)
	if err != nil {
		return nil, err
	}

	key := filepath.Join(namespace, name)
	rev := clientv3.CreateRevision(key)
	cmp := clientv3.Compare(rev, "=", 0)
	then := clientv3.OpPut(key, addr, clientv3.WithLease(resp.ID))
	txn, err := d.cli.Txn(context.TODO()).If(cmp).Then(then).Commit()
	if err != nil {
		return nil, err
	}
	if !txn.Succeeded {
		return nil, errors.New("service exists")
	}
	return d.keepAlive(resp.ID), nil
}

func (d *Discoverer) keepAlive(id clientv3.LeaseID) api.KeepAliveFunc {
	return func() error {
		_, err := d.cli.KeepAliveOnce(context.TODO(), id)
		return err
	}
}

// Unregister service
func (d *Discoverer) Unregister(namespace, name string) error {
	key := filepath.Join(namespace, name)
	_, err := d.cli.Delete(context.TODO(), key)
	return err
}

// Watch services, returns event and cancel func
func (d *Discoverer) Watch(namespace string) (<-chan *api.Event, api.CancelFunc, error) {
	event := make(chan *api.Event, 1024)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer close(event)

		wc := d.cli.Watch(ctx, namespace, clientv3.WithPrefix())
		for {
			select {
			case <-ctx.Done():
			case resp := <-wc:
				for _, ev := range resp.Events {
					action := ev.Type.String()
					event <- &api.Event{
						Action: strings.ToLower(action),
						Name:   string(ev.Kv.Key),
						Addr:   string(ev.Kv.Value),
					}
				}
			}
		}
	}()
	return event, api.CancelFunc(cancel), nil
}

// Services list all service
func (d *Discoverer) Services(namespace string) (map[string]string, error) {
	resp, err := d.cli.Get(context.TODO(), namespace, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	svrs := make(map[string]string)
	for _, kv := range resp.Kvs {
		svrs[string(kv.Key)] = string(kv.Value)
	}
	return svrs, nil
}
