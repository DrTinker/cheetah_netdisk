package trans

import (
	"NetDesk/client"
	"NetDesk/conf"
	"NetDesk/service"

	"github.com/sirupsen/logrus"
)

func TransferObjectHandler() {
	// 初始化channel
	setting, err := client.GetMQClient().InitTransfer(conf.Exchange, conf.Routing_Key)
	if err != nil {
		logrus.Error("[TransferObjectHandler] init channel error: %v", err)
	}
	err = client.GetMQClient().Consume(setting, conf.Transfer_COS_Queue, "transfer_consumer", service.UploadConsumerMsg)
	if err != nil {
		logrus.Error("[TransferObjectHandler] init channel error: %v", err)
	}
}
