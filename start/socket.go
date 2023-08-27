package start

import (
	"NetDesk/client"
	"NetDesk/infrastructure/socket"
)

func InitSocket() {
	impl := socket.NewSocketClientImpl()

	client.InitSocketClient(impl)
}
