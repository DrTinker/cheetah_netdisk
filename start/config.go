package start

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/infrastructure/config"
)

// 加载启动项
func InitConfig() {
	impl := config.NewConfigClientImpl()
	err := impl.Load(conf.App)
	if err != nil {
		panic(err)
	}

	client.InitConfigClient(impl)
}
