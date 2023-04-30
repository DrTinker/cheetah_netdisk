package client

import (
	"sync"
	"time"
)

type CacheClient interface {
	// 通用能力
	Exists(key string) (num int64, err error)
	Expire(key string, expire time.Duration) error
	// string
	Get(key string) (res string, err error)
	Set(key, val string) error
	SetWithExpire(key, val string, expire time.Duration) error
	DelBatch(keys string) (num int64, err error)
	// hset
	HSet(key, filed string, val interface{}) error
	HMSet(key string, fileds map[string]interface{}) error
	HGet(key, filed string) (string, error)
	HGetAll(key string) (map[string]string, error)
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
