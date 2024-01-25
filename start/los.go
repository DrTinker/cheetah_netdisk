package start

import (
	"NetDisk/client"
	"NetDisk/infrastructure/los"
)

func InitLOS() {
	cfg, err := client.GetConfigClient().GetLOSConfig()
	if err != nil {
		panic(err)
	}
	impl, err := los.NewLOSClientImpl(cfg)
	if err != nil {
		panic(err)
	}

	client.InitLOSClientt(impl)
}
