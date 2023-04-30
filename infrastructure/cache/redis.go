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

// 判断是否存在
func (c *CacheClientImpl) Exists(key string) (num int64, err error) {
	return c.CacheClient.Exists(key).Result()
}

// 设置过期时间
func (c *CacheClientImpl) Expire(key string, expire time.Duration) error {
	return c.CacheClient.Expire(key, expire).Err()
}

// string get
func (c *CacheClientImpl) Get(key string) (res string, err error) {
	return c.CacheClient.Get(key).Result()
}

// string set
func (c *CacheClientImpl) Set(key, val string) error {
	return c.CacheClient.Set(key, val, 0).Err()
}

// string set with expire
func (c *CacheClientImpl) SetWithExpire(key, val string, expire time.Duration) error {
	return c.CacheClient.Set(key, val, expire).Err()
}

// 批量删除keys
func (c *CacheClientImpl) DelBatch(keys string) (num int64, err error) {
	return c.CacheClient.Del(keys).Result()
}

// hset set
func (c *CacheClientImpl) HSet(key, filed string, val interface{}) error {
	return c.CacheClient.HSet(key, filed, val).Err()
}

// hset set多个
func (c *CacheClientImpl) HMSet(key string, fileds map[string]interface{}) error {
	return c.CacheClient.HMSet(key, fileds).Err()
}

// hset get
func (c *CacheClientImpl) HGet(key, filed string) (string, error) {
	return c.CacheClient.HGet(key, filed).Result()
}

// hset get all
func (c *CacheClientImpl) HGetAll(key string) (map[string]string, error) {
	return c.CacheClient.HGetAll(key).Result()
}
