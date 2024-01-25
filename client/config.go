package client

import (
	"NetDisk/models"
	"sync"
)

type ConfigClient interface {
	Load(path string) error
	GetHttpConfig() (*models.HttpConfig, error)
	GetDBConfig() (driver, source string, err error)
	GetEmailConfig() (*models.EmailConfig, error)
	GetCacheConfig() (addr, pwd string, err error)
	GetCOSConfig() (*models.COSConfig, error)
	GetLocalConfig() (*models.LocalConfig, error)
	GetMQConfig() (*models.MQConfig, error)
	GetLOSConfig() (*models.LOSConfig, error)
}

var (
	configClient ConfigClient
	configOnce   sync.Once
)

func GetConfigClient() ConfigClient {
	return configClient
}

func InitConfigClient(client ConfigClient) {
	configOnce.Do(
		func() {
			configClient = client
		},
	)
}
