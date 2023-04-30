package client

import (
	"NetDesk/common/models"
	"sync"
)

type MQClient interface {
	InitTransfer(exchange, key string) (*models.TransferSetting, error)
	Publish(setting *models.TransferSetting, msg []byte) error
	Consume(setting *models.TransferSetting, queue, consumer string, callback func(msg []byte) bool) error
}

var (
	mq     MQClient
	MQOnce sync.Once
)

func GetMQClient() MQClient {
	return mq
}

func InitMQClient(client MQClient) {
	MQOnce.Do(
		func() {
			mq = client
		},
	)
}
