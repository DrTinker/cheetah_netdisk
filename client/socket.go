package client

import (
	"net/http"
	"sync"
)

type SocketClient interface {
	AddConn(w http.ResponseWriter, r *http.Request, id string) error
	SendMsg(id string, msg interface{}) error
	ReadMsg(id string, msg interface{}) error
	DeleteConn(id string)
	CheckConn(id string) bool
}

var (
	socket     SocketClient
	SocketOnce sync.Once
)

func GetSocketClient() SocketClient {
	return socket
}

func InitSocketClient(client SocketClient) {
	SocketOnce.Do(
		func() {
			socket = client
		},
	)
}
