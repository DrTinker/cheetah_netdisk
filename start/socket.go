package start

import (
	"NetDisk/client"
	"NetDisk/infrastructure/socket"
)

func InitSocket() {
	impl := socket.NewSocketClientImpl()

	client.InitSocketClient(impl)
}
