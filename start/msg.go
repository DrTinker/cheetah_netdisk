package start

import (
	"NetDesk/client"
	"NetDesk/infrastructure/msg"
)

func InitMsg() {
	impl, err := msg.NewMsgClientImpl()
	if err != nil {
		panic(err)
	}

	client.InitMsglient(impl)
}
