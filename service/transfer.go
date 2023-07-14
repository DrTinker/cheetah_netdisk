package service

import (
	"NetDesk/client"
	"NetDesk/conf"
	"NetDesk/helper"
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
		log.Error("[UploadConsumerMsg] parse msg error: ", err)
		return false
	}
	// 根据msg读取本地文件上传cos
	err = client.GetCOSClient().UpLoadLocalFile(data.Des, data.Src)
	if err != nil {
		log.Error("[UploadConsumerMsg] upload cos error: ", err)
		return false
	}
	// 修改数据表
	err = client.GetDBClient().UpdateFileStoreTypeByHash(data.FileHash, data.StoreType)
	if err != nil {
		log.Error("[UploadConsumerMsg] update db error: ", err)
		return false
	}
	err = client.GetDBClient().UpdateTransState(data.UploadID, conf.Trans_Success)
	if err != nil {
		log.Error("[UploadConsumerMsg] update db error: ", err)
		return false
	}
	// 删除tmp下文件
	err = helper.DelFile(data.Src, 0)
	if err != nil {
		log.Error("[UploadConsumerMsg] remove tmp file error: ", err)
		return false
	}
	log.Info("[UploadConsumerMsg] transfer file ", data.Src, " success")
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
	log.Info("[UploadProduceMsg] send msg ", data.Src, " success")
	return nil
}
