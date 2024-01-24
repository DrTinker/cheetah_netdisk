package start

import (
	"NetDisk/client"
	"NetDisk/infrastructure/cache"
)

func InitCache() {
	addr, pwd, err := client.GetConfigClient().GetCacheConfig()
	if err != nil {
		panic(err)
	}
	impl, err := cache.NewCacheClientImpl(addr, pwd)
	if err != nil {
		panic(err)
	}
	client.InitCacheClient(impl)
}
