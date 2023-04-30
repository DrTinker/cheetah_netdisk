package start

import (
	"NetDesk/common/client"
	"NetDesk/common/infrastructure/mq"
	"fmt"
)

func InitMQ() {
	cfg, err := client.GetConfigClient().GetMQConfig()
	if err != nil {
		panic(err)
	}
	url := fmt.Sprintf("%s://%s:%s@%s:%d/", cfg.Proto, cfg.User, cfg.Pwd, cfg.Address, cfg.Port)
	impl, err := mq.NewMQClientImpl(url)
	if err != nil {
		panic(err)
	}

	client.InitMQClient(impl)
}
