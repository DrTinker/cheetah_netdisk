package client

import (
	"sync"
	"time"
)

type CacheClient interface {
	// 通用能力
	Exists(key string) (num int64, err error)
	Get(key string) (res string, err error)
	Set(key, val string) error
	SetWithExpire(key, val string, expire time.Duration) error
	DelBatch(keys string) (num int64, err error)
}

var (
	cache     CacheClient
	CacheOnce sync.Once
)

func GetCacheClient() CacheClient {
	return cache
}

func InitCacheClient(client CacheClient) {
	CacheOnce.Do(
		func() {
			cache = client
		},
	)
}
