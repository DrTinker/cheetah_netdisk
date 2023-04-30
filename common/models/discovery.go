package models

import (
	"net/http"
	"time"
)

type DiscoveryParam struct {
	Effect bool // 判断配置是否生效，为false使用默认配置
	// Address is the address of the Consul server
	Address string

	// Scheme is the URI scheme for the Consul server
	Scheme string

	// Prefix for URIs for when consul is behind an API gateway (reverse
	// proxy).  The API gateway must strip off the PathPrefix before
	// passing the request onto consul.
	PathPrefix string

	// Datacenter to use. If not provided, the default agent datacenter is used.
	Datacenter string

	// Transport is the Transport to use for the http client.
	Transport *http.Transport

	// HttpClient is the client to use. Default will be
	// used if not provided.
	HttpClient *http.Client

	// WaitTime limits how long a Watch will block. If not provided,
	// the agent default values will be used.
	WaitTime time.Duration

	// Token is used to provide a per-request ACL token
	// which overrides the agent's default token.
	Token string

	// TokenFile is a file containing the current token to use for this client.
	// If provided it is read once at startup and never again.
	TokenFile string

	// Namespace is the name of the namespace to send along for the request
	// when no other Namespace is present in the QueryOptions
	Namespace string

	// Partition is the name of the partition to send along for the request
	// when no other Partition is present in the QueryOptions
	Partition string
}
