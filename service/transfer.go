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
	err = client.GetCOSClient().UpLoadLocalFile(data.FileKey, data.TmpPath)
	if err != nil {
		log.Error("[UploadConsumerMsg] upload cos error: ", err)
		return false
	}
	// 读取缩略图上传
	err = client.GetCOSClient().UpLoadLocalFile(data.TnFileKey, data.Thumbnail)
	// 没有缩略图只提示
	if err != nil {
		log.Warn("[UploadConsumerMsg] upload thumbnail cos error: ", data.FileKey, err)
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
	err = helper.DelFile(data.TmpPath)
	if err != nil {
		log.Error("[UploadConsumerMsg] remove tmp file error: ", err)
		return false
	}
	// 删除tmp下缩略图，错误只提示不返回
	err = helper.DelFile(data.Thumbnail)
	if err != nil {
		log.Warn("[UploadConsumerMsg] remove thumbnail error: ", data.FileKey, err)
	}
	log.Info("[UploadConsumerMsg] transfer file ", data.TmpPath, " success")
	return true
}

func UploadProduceMsg(data *models.TransferMsg) error {
	// TODO rabbit 不可用问题研究
	setting, err := client.GetMQClient().InitTransfer(conf.Exchange, conf.Routing_Key)
	defer client.GetMQClient().ReleaseChannel(setting)
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
	log.Info("[UploadProduceMsg] send msg ", data.TmpPath, " success")
	return nil
}
