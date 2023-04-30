package start

import (
	"NetDesk/common/client"
	"NetDesk/common/infrastructure/cos"
)

func InitCOS() {
	cfg, err := client.GetConfigClient().GetCOSConfig()
	if err != nil {
		panic(err)
	}
	impl, err := cos.NewCOSClientImpl(cfg)
	if err != nil {
		panic(err)
	}

	client.InitCOSClient(impl)
}
