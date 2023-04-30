package start

import (
	"NetDesk/common/client"
	"NetDesk/common/infrastructure/msg"
)

func InitMsg() {
	impl, err := msg.NewMsgClientImpl()
	if err != nil {
		panic(err)
	}

	client.InitMsglient(impl)
}
