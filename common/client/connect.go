package client

import (
	"NetDesk/common/conf"
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type RpcConnect struct {
	Target string
}

func NewRpcConnect(name string) (*RpcConnect, error) {
	cfg, err := GetConfigClient().GetDiscoveryConfig()
	if err != nil {
		return nil, errors.Wrap(err, "[RpcConnect] get rpc config err ")
	}
	return &RpcConnect{
		Target: fmt.Sprintf("%s://%s:%d/%s", cfg.Tool, cfg.Address, cfg.Port, name),
	}, nil
}

func (r *RpcConnect) GetConnect() (*grpc.ClientConn, error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(r.Target, grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("[RpcConnect] [%s] grpc dial ", conf.User_Service_Name))
	}
	return conn, nil
}
