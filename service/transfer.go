package service

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"encoding/json"
	"fmt"

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
	// 弃用，目前为从COS直接下载
	if data.Task == conf.DownloadMod {
		key := helper.GenDownloadPartInfoKey(data.TransID)
		// 查看文件是否已经存在
		exist, _ := helper.PathExists(data.TmpPath)
		// 不存在则下载
		if !exist {
			err = client.GetCOSClient().DownloadLocal(data.FileKey, data.TmpPath)
			if err != nil {
				log.Error("[TransferConsumerMsg] parse msg error: ", err)
				// 更新状态为abort
				client.GetCacheClient().HSet(key, conf.DownloadPartReadyKey, conf.DownloadReadyAbort)
				return false
			}
		}
		// 存在直接更新
		// 更新状态为done
		client.GetCacheClient().HSet(key, conf.DownloadPartReadyKey, conf.DownloadReadyDone)
		log.Info("[TransferConsumerMsg] transfer file ", data.TmpPath, " success")
		return true
	}
	// 上传任务
	// 拼接磁盘临时地址
	cfg, err := client.GetConfigClient().GetLocalConfig()
	if err != nil {
		log.Error("[TransferConsumerMsg] get loacl config error: ", err)
		return false
	}
	filePath, tnPath := fmt.Sprintf("%s/%s", cfg.TmpPath, data.FileName), fmt.Sprintf("%s/%s", cfg.TmpPath, data.TnName)
	// 删除磁盘临时文件
	defer func() {
		helper.DelFile(filePath)
		helper.DelFile(tnPath)
	}()

	// 先处理缩略图，减少用户延迟感知
	// 下载缩略图, 不处理缩略图缺失
	client.GetLOSClient().FGetObject(data.Thumbnail, tnPath)
	// 读取缩略图上传, 不处理缩略图缺失
	client.GetCOSClient().UpLoadLocalFile(data.TnFileKey, tnPath)

	// 后处理文件
	// 从los下载文件到本地磁盘
	err = client.GetLOSClient().FGetObject(data.TmpPath, filePath)
	if err != nil {
		log.Error("[TransferConsumerMsg] get file from los error: ", err)
		return false
	}
	// 根据msg读取本地文件上传cos
	err = client.GetCOSClient().UpLoadLocalFile(data.FileKey, filePath)
	if err != nil {
		log.Error("[TransferConsumerMsg] upload cos error: ", err)
		return false
	}

	// 修改数据表
	err = client.GetDBClient().UpdateFileStoreTypeByHash(data.FileHash, data.StoreType)
	if err != nil {
		log.Error("[TransferConsumerMsg] update db error: ", err)
		return false
	}
	err = client.GetDBClient().UpdateTransState(data.TransID, conf.TransSuccess)
	if err != nil {
		log.Error("[TransferConsumerMsg] update db error: ", err)
		return false
	}

	log.Info("[TransferConsumerMsg] transfer file ", data.TmpPath, " success")
	return true
}

func TransferProduceMsg(data *models.TransferMsg) error {
	// TODO rabbit 不可用问题研究
	setting, err := client.GetMQClient().InitTransfer(conf.Exchange, conf.RoutingKey)
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
