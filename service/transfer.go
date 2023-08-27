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
func TransferConsumerMsg(msg []byte) bool {
	// 解析消息
	data := &models.TransferMsg{}
	err := json.Unmarshal(msg, data)
	if err != nil {
		log.Error("[TransferConsumerMsg] parse msg error: ", err)
		return false
	}
	// 判断任务类型
	// 下载任务
	if data.Task == conf.Download_Mod {
		key := helper.GenDownloadPartInfoKey(data.TransID)
		// 查看文件是否已经存在
		exist, _ := helper.PathExists(data.TmpPath)
		// 不存在则下载
		if !exist {
			err = client.GetCOSClient().DownloadLocal(data.FileKey, data.TmpPath)
			if err != nil {
				log.Error("[TransferConsumerMsg] parse msg error: ", err)
				// 更新状态为abort
				client.GetCacheClient().HSet(key, conf.Download_Part_Ready_Key, conf.Download_Ready_Abort)
				return false
			}
		}
		// 存在直接更新
		// 更新状态为done
		client.GetCacheClient().HSet(key, conf.Download_Part_Ready_Key, conf.Download_Ready_Done)
		log.Info("[TransferConsumerMsg] transfer file ", data.TmpPath, " success")
		return true
	}
	// 上传任务
	// 根据msg读取本地文件上传cos
	err = client.GetCOSClient().UpLoadLocalFile(data.FileKey, data.TmpPath)
	if err != nil {
		log.Error("[TransferConsumerMsg] upload cos error: ", err)
		return false
	}
	// 读取缩略图上传
	err = client.GetCOSClient().UpLoadLocalFile(data.TnFileKey, data.Thumbnail)
	// 没有缩略图只提示
	if err != nil {
		log.Warn("[TransferConsumerMsg] upload thumbnail cos error: ", data.FileKey, err)
	}

	// 修改数据表
	err = client.GetDBClient().UpdateFileStoreTypeByHash(data.FileHash, data.StoreType)
	if err != nil {
		log.Error("[TransferConsumerMsg] update db error: ", err)
		return false
	}
	err = client.GetDBClient().UpdateTransState(data.TransID, conf.Trans_Success)
	if err != nil {
		log.Error("[TransferConsumerMsg] update db error: ", err)
		return false
	}
	// 删除tmp下文件
	err = helper.DelFile(data.TmpPath)
	if err != nil {
		log.Error("[TransferConsumerMsg] remove tmp file error: ", err)
		return false
	}
	// 删除tmp下缩略图，错误只提示不返回
	err = helper.DelFile(data.Thumbnail)
	if err != nil {
		log.Warn("[TransferConsumerMsg] remove thumbnail error: ", data.FileKey, err)
	}
	log.Info("[TransferConsumerMsg] transfer file ", data.TmpPath, " success")
	return true
}

func TransferProduceMsg(data *models.TransferMsg) error {
	// TODO rabbit 不可用问题研究
	setting, err := client.GetMQClient().InitTransfer(conf.Exchange, conf.Routing_Key)
	defer client.GetMQClient().ReleaseChannel(setting)
	if err != nil {
		return errors.Wrap(err, "[UploadObject] init transfer channel error: ")
	}
	// 序列化msg
	msg, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "[UploadObject] parse msg error: ")
	}
	// 发布
	err = client.GetMQClient().Publish(setting, msg)
	if err != nil {
		return errors.Wrap(err, "[UploadObject] publish msg error: ")
	}
	log.Info("[UploadProduceMsg] send msg ", data.TmpPath, " success")
	return nil
}
