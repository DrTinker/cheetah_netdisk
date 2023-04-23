package service

import (
	"NetDesk/client"
	"NetDesk/conf"
	"NetDesk/models"
	"encoding/json"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// 消息队列消费逻辑
func UploadConsumerMsg(msg []byte) bool {
	// 解析消息
	data := &models.TransferMsg{}
	err := json.Unmarshal(msg, data)
	if err != nil {
		log.Error("[UploadConsumerMsg] parse msg error: %v", err)
		return false
	}
	// 根据msg读取本地文件上传cos
	err = client.GetCOSClient().UpLoadLocalFile(data.Des, data.Src)
	if err != nil {
		log.Error("[UploadConsumerMsg] upload cos error: %v", err)
		return false
	}
	// 修改数据表
	err = client.GetDBClient().UpdateFileStoreTypeByHash(data.FileHash, data.StoreType)
	if err != nil {
		log.Error("[UploadConsumerMsg] update db error: %v", err)
		return false
	}
	return true
}

func UploadProduceMsg(data *models.TransferMsg) error {
	setting, err := client.GetMQClient().InitTransfer(conf.Exchange, conf.Routing_Key)
	if err != nil {
		return errors.Wrap(err, "[UploadObject] init transfer channel error: ")
	}
	msg, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "[UploadObject] parse msg error: ")
	}
	err = client.GetMQClient().Publish(setting, msg)
	if err != nil {
		return errors.Wrap(err, "[UploadObject] publish msg error: ")
	}
	return nil
}
