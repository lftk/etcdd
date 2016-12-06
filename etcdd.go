package etcdd

import (
	"github.com/4396/etcdd/api"
	"github.com/4396/etcdd/etcddv2"
	"github.com/4396/etcdd/etcddv3"
)

// NewV2 returns v2 discoverer
func NewV2(endpoints []string) (api.Discoverer, error) {
	return etcddv2.New(endpoints)
}

// NewV3 returns v3 discoverer
func NewV3(endpoints []string) (api.Discoverer, error) {
	return etcddv3.New(endpoints)
}

// NewV2WithAuth need username and password
func NewV2WithAuth(endpoints []string, username, password string) (api.Discoverer, error) {
	return etcddv2.NewWithAuth(endpoints, username, password)
}

// NewV3WithAuth need username and password
func NewV3WithAuth(endpoints []string, username, password string) (api.Discoverer, error) {
	return etcddv3.NewWithAuth(endpoints, username, password)
}
