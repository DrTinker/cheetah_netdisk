package client

import (
	"NetDisk/models"
	"sync"
)

type MQClient interface {
	Publish(setting *models.TransferSetting, msg []byte) error
	Consume(setting *models.TransferSetting, queue, consumer string, callback func(msg []byte) bool) error
	KeepAlive()
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
