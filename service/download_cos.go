package service

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"strconv"

	"github.com/pkg/errors"
)

func InitDownloadCOS(param *models.DownloadObjectParam) (*models.InitDownloadResult, error) {
	// 接收参数
	downloadID := param.DownloadID
	UserFileUuid := param.UserFileUuid
	// 查询file信息
	user_file, err := client.GetDBClient().GetUserFileByUuid(UserFileUuid)
	if err != nil {
		return nil, err
	}
	// 检查下载文件的人是否是文件持有者
	if user_file.UserUuid != param.UserUuid {
		return nil, conf.InvaildOwnerError
	}
	// 通过uuid查询文件信息
	fileKey, err := client.GetDBClient().GetFileKeyByUserFileUuid(UserFileUuid)
	if err != nil {
		return nil, err
	}
	size := user_file.Size
	// 获取预签名
	url, err := client.GetCOSClient().GetPresignedUrl(fileKey, conf.DefaultSignExpire)
	if err != nil {
		return nil, err
	}
	// 读取配置完善url
	cfg, err := client.GetConfigClient().GetCOSConfig()
	if err != nil {
		return nil, err
	}
	url = cfg.Domain + url
	// 返回参数
	res := &models.InitDownloadResult{}
	// redis key
	infoKey := helper.GenDownloadPartInfoKey(downloadID)
	// 分片数量
	count := size/conf.FilePartSizeMax + 1
	// 分片列表
	var chunkList []int

	// 断点续传
	// 尝试获取分片信息，如果存在则说明之前上传过，触发断点续传逻辑
	if param.Continue {
		tmpInfo, err := client.GetCacheClient().HGetAll(infoKey)
		if len(tmpInfo) != 0 && err == nil {
			// 记录已经下载的分片
			for k, _ := range tmpInfo {
				if i, err := strconv.Atoi(k); err == nil {
					chunkList = append(chunkList, i)
				}
			}
			res.ChunkCount = count
			res.ChunkList = chunkList
			res.DownloadID = tmpInfo[conf.DownloadPartInfoKey]
			res.Hash = user_file.Hash
			res.Url = url
			return res, nil
		}
	}
	// 首次下载，先将COS文件下载到服务器
	// 生成分块下载信息
	info := map[string]interface{}{
		conf.DownloadPartInfoKey:       downloadID,
		conf.DownloadPartInfoCSizeKey:  conf.FilePartSizeMax,
		conf.DownloadPartInfoCCountKey: count,
		conf.DownloadPartFileSizeKey:   size,
	}
	// 写redis
	err = client.GetCacheClient().HMSet(infoKey, info)
	if err != nil {
		return nil, errors.Wrap(err, "[InitDownload] set download info error: ")
	}
	// 设置过期时间
	err = client.GetCacheClient().Expire(infoKey, conf.Trans_Part_Slice_Expire)
	if err != nil {
		return nil, errors.Wrap(err, "[InitDownload] set download info error: ")
	}

	// 未下载过或者redis中key已过期
	// redis过期的情况在GetTransList接口中已经处理
	// 即用户已进入trans页面就会将db中过期的记录status改为失败
	// 调用此接口时，不论之前是nil还是fail都应更改状态为process
	// 创建trans记录
	trans := &models.Trans{
		Uuid:         downloadID,
		UserUuid:     param.UserUuid,
		UserFileUuid: param.UserFileUuid,
		ParentUuid:   param.ParentUuid,
		FileKey:      fileKey,
		Hash:         user_file.Hash,
		LocalPath:    param.LocalPath,
		Size:         user_file.Size,
		Name:         user_file.Name,
		Ext:          user_file.Ext,
		Status:       conf.TransProcess,
		Isdown:       conf.DownloadMod,
	}
	err = client.GetDBClient().CreateTrans(trans)
	if err != nil {
		return nil, errors.Wrap(err, "[InitDownload] set trans record error: ")
	}

	// 返回下载签名
	res.ChunkCount = count
	res.ChunkList = chunkList
	res.DownloadID = downloadID
	res.Hash = user_file.Hash
	res.Url = url

	return res, nil
}

// 写入分片下载信息，不进行实际下载，实际下载由客户端直接请求COS
func DownloadPartCOS(downloadID string, chunkNum int) error {
	key := helper.GenDownloadPartInfoKey(downloadID)
	// 读取redis
	num, err := client.GetCacheClient().Exists(key)
	if err != nil {
		return errors.Wrap(err, "[DownloadPart] get download info err: ")
	}
	if num == 0 {
		return errors.New("[DownloadPart] get download info err: empty key")
	}
	// 更新redis
	err = client.GetCacheClient().HSet(key, strconv.Itoa(chunkNum), strconv.Itoa(chunkNum))
	if err != nil {
		return errors.Wrap(err, "[DownloadPart] update cache error: ")
	}

	return nil
}

// 分块下载完成
func CompleteDownloadPartCOS(downloadID string) error {
	// 查看redis中记录是否完整
	infoKey := helper.GenDownloadPartInfoKey(downloadID)
	infoMap, err := client.GetCacheClient().HGetAll(infoKey)
	if err != nil {
		return errors.Wrap(err, "[CompleteDownloadPart] get download info error: ")
	}
	if _, ok := infoMap[conf.DownloadPartInfoCCountKey]; !ok {
		return errors.Wrap(conf.MapNotHasError, "[CompleteDownloadPart] get chunk count error: ")
	}
	// 忽略错误
	count, _ := strconv.Atoi(infoMap[conf.DownloadPartInfoCCountKey])
	// 除去info固定的n个，剩下的fields都对应一个已经上传的分片
	// 如果分片不完整，则返回错误
	if (count) != len(infoMap)-conf.DownloadPartCOSInfoFileds {
		return errors.Wrap(conf.ChunkMissError, "[CompleteUploadPart] unable to complete: ")
	}
	// 删除rediskey
	client.GetCacheClient().DelBatch(infoKey)
	// 更改trans表记录状态
	err = client.GetDBClient().UpdateTransState(downloadID, conf.TransSuccess)
	if err != nil {
		return err
	}

	return nil
}

// 获取COS签名
func DownloadTotal(param *models.DownloadObjectParam) (string, error) {
	// 接收参数
	UserUuid := param.UserUuid
	UserFileUuid := param.UserFileUuid
	// downloadID := helper.GenDownloadID(UserUuid, UserFileUuid)
	// 获取fileKey
	fileKey, err := client.GetDBClient().GetFileKeyByUserFileUuid(UserFileUuid)
	if err != nil {
		return "", err
	}
	// 查询file信息
	user_file, err := client.GetDBClient().GetUserFileByUuid(UserFileUuid)
	if err != nil {
		return "", err
	}
	// 获取预签名
	url, err := client.GetCOSClient().GetPresignedUrl(fileKey, conf.DefaultSignExpire)
	if err != nil {
		return "", err
	}
	// 读取配置完善url
	cfg, err := client.GetConfigClient().GetCOSConfig()
	if err != nil {
		return "", err
	}
	url = cfg.Domain + url
	// 创建trans记录
	trans := &models.Trans{
		Uuid:         param.DownloadID,
		UserUuid:     UserUuid,
		UserFileUuid: param.UserFileUuid,
		ParentUuid:   param.ParentUuid,
		FileKey:      fileKey,
		Hash:         user_file.Hash,
		LocalPath:    param.LocalPath,
		Size:         user_file.Size,
		Name:         user_file.Name,
		Ext:          user_file.Ext,
		Status:       conf.TransSuccess,
		Isdown:       conf.DownloadMod,
	}
	err = client.GetDBClient().CreateTrans(trans)
	if err != nil {
		return "", errors.Wrap(err, "[InitDownload] set trans record error: ")
	}
	// 返回
	return url, nil
}

func CancelDownload(downloadID string) error {
	// redis查看是否存在记录，若不存在则一定不在进行中
	infoKey := helper.GenDownloadPartInfoKey(downloadID)
	num, err := client.GetCacheClient().Exists(infoKey)
	if err != nil {
		return errors.Wrap(err, "[CancelDownload] get download info err ")
	}
	// 不存在直接说明已经结束或者失败
	if num == 0 {
		return errors.Wrapf(conf.TransFinishError, "[CancelDownload] download: %s is finished ", downloadID)
	}
	// 存在说明正在进行
	// 删除trans表记录
	err = client.GetDBClient().DelTransByUuid(downloadID)
	if err != nil {
		return errors.Wrap(err, "[CancelUpload] delete trans error: ")
	}
	// 删除redis缓存
	client.GetCacheClient().DelBatch(infoKey)
	return nil
}
