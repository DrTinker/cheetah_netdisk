package trans

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/models"
	"NetDisk/service"

	"github.com/sirupsen/logrus"
)

func TransferObjectHandler() {
	setting := &models.TransferSetting{
		Exchange:  conf.Exchange,
		RoutinKey: conf.RoutingKey,
	}
	err := client.GetMQClient().Consume(setting, conf.TransferCOSQueue, "transfer_consumer", service.TransferConsumerMsg)
	if err != nil {
		logrus.Error("[TransferObjectHandler] init channel error: ", err)
	}
}
