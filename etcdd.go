package discovery

import (
	"github.com/4396/etcd-discovery/api"
	"github.com/4396/etcd-discovery/etcddv2"
	"github.com/4396/etcd-discovery/etcddv3"
)

func NewV2(endpoints []string) (api.Discoverer, error) {
	return etcddv2.New(endpoints)
}

func NewV3(endpoints []string) (api.Discoverer, error) {
	return etcddv3.New(endpoints)
}
