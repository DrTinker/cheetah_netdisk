package start

import (
	"NetDesk/common/client"
	"NetDesk/common/conf"
	"NetDesk/common/infrastructure/config"
)

// 加载启动项
func InitConfig() {
	impl := config.NewConfigClientImpl()
	err := impl.Load(conf.AppCfg)
	if err != nil {
		panic(err)
	}

	client.InitConfigClient(impl)
}
