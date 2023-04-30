package start

import (
	"NetDesk/common/client"
	"NetDesk/common/infrastructure/consul"
	"NetDesk/common/models"
)

func InitDiscoveryClient() {
	impl, err := consul.NewDiscoveryClientImpl(&models.DiscoveryParam{Effect: false})
	if err != nil {
		panic(err)
	}
	client.InitDiscoveryClient(impl)
}
