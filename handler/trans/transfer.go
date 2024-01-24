package trans

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/service"

	"github.com/sirupsen/logrus"
)

func TransferObjectHandler() {
	// 初始化channel
	setting, err := client.GetMQClient().InitTransfer(conf.Exchange, conf.RoutingKey)
	if err != nil {
		logrus.Error("[TransferObjectHandler] init channel error: ", err)
	}
	err = client.GetMQClient().Consume(setting, conf.TransferCOSQueue, "transfer_consumer", service.TransferConsumerMsg)
	if err != nil {
		logrus.Error("[TransferObjectHandler] init channel error: ", err)
	}
}
