package cache

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type CacheClientImpl struct {
	CacheClient *redis.Client
}

func NewCacheClientImpl(addr, pwd string) (*CacheClientImpl, error) {
	rdb := &CacheClientImpl{}
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       0,
	})
	pong, err := c.Ping().Result()
	if err != nil {
		return nil, err
	}
	logrus.Info("redis init success: ", pong)
	rdb.CacheClient = c
	return rdb, nil
}

func (c *CacheClientImpl) Exists(key string) (num int64, err error) {
	return c.CacheClient.Exists(key).Result()
}

func (c *CacheClientImpl) Get(key string) (res string, err error) {
	return c.CacheClient.Get(key).Result()
}
func (c *CacheClientImpl) Set(key, val string) error {
	return c.CacheClient.Set(key, val, 0).Err()
}
func (c *CacheClientImpl) SetWithExpire(key, val string, expire time.Duration) error {
	return c.CacheClient.Set(key, val, expire).Err()
}

func (c *CacheClientImpl) DelBatch(keys string) (num int64, err error) {
	return c.CacheClient.Del(keys).Result()
}
