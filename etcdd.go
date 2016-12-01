package etcdd

import (
	"github.com/4396/etcdd/api"
	"github.com/4396/etcdd/etcddv2"
	"github.com/4396/etcdd/etcddv3"
)

func NewV2(endpoints []string) (api.Discoverer, error) {
	return etcddv2.New(endpoints)
}

func NewV3(endpoints []string) (api.Discoverer, error) {
	return etcddv3.New(endpoints)
}
