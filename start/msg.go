package start

import (
	"NetDisk/client"
	"NetDisk/infrastructure/msg"
)

func InitMsg() {
	impl, err := msg.NewMsgClientImpl()
	if err != nil {
		panic(err)
	}

	client.InitMsgClient(impl)
}
