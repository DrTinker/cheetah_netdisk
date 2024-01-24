package start

import (
	"NetDisk/client"
	"NetDisk/infrastructure/mq"
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

	impl.KeepAlive()
	client.InitMQClient(impl)
}
