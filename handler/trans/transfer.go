package trans

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/models"
	"NetDisk/service"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
)

func TransferObjectHandler() {
	setting := &models.TransferSetting{
		Exchange:  conf.Exchange,
		RoutinKey: conf.RoutingKey,
	}
	for {
		err := client.GetMQClient().Consume(setting, conf.TransferCOSQueue, "transfer_consumer", service.TransferConsumerMsg)
		if err != nil {
			if errors.Is(err, conf.MQConnectionClosedError) {
				// 连接问题则每隔10s尝试
				logrus.Error("[TransferObjectHandler] conn error: ", err)
				time.Sleep(10 * time.Second)
				continue
			}
			logrus.Error("[TransferObjectHandler] consume error: ", err)
		}
	}
}
