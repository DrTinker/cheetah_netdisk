package client

import (
	"sync"
)

type DiscoveryClient interface {
	RegisterService(name, id, host string, port int) error
	KeepAlive(id string)
}

var (
	discoveryClient DiscoveryClient
	DiscoveryOnce   sync.Once
)

func GetDiscoveryClient() DiscoveryClient {
	return discoveryClient
}

func InitDiscoveryClient(client DiscoveryClient) {
	DiscoveryOnce.Do(
		func() {
			discoveryClient = client
		},
	)
}
