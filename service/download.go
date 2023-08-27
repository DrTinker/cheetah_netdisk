package service

import (
	"NetDesk/client"
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

// 初始化传输，返回上传签名，并更新redis和db的值
func InitDownload(param *models.DownloadObjectParam) (*models.InitDownloadResult, error) {
	// 接收参数
	downloadID := param.DownloadID
	user_file_uuid := param.User_File_Uuid
	// 查询file信息
	user_file, err := client.GetDBClient().GetUserFileByUuid(user_file_uuid)
	if err != nil {
		return nil, err
	}
	// 通过uuid查询文件信息
	fileKey, err := client.GetDBClient().GetFileKeyByUserFileUuid(user_file_uuid)
	if err != nil {
		return nil, err
	}
	// 检查下载文件的人是否是文件持有者
	if user_file.User_Uuid != param.User_Uuid {
		return nil, conf.InvaildOwnerError
	}
	// 生成filePath, 文件再本地暂存路径
	hash, ext := user_file.Hash, user_file.Ext
	cfg, err := client.GetConfigClient().GetLocalConfig()
	if err != nil {
		return nil, errors.Wrap(err, "[DownloadFile] parse filePath err ")
	}
	filePath := fmt.Sprintf("%s/%s.%s", cfg.FilePath, hash, ext)
	size := user_file.Size
	// 返回参数
	res := &models.InitDownloadResult{}
	// redis key
	infoKey := helper.GenDownloadPartInfoKey(downloadID)
	// 分片数量
	count := size/conf.File_Part_Size_Max + 1
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
			res.DownloadID = tmpInfo[conf.Download_Part_Info_Key]
			res.Hash = hash
			return res, nil
		}
	}
	// 首次下载，先将COS文件下载到服务器
	// 生成分块下载信息
	info := map[string]interface{}{
		conf.Download_Part_Info_Key:        downloadID,
		conf.Download_Part_Info_CSize_Key:  conf.File_Part_Size_Max,
		conf.Download_Part_Info_CCount_Key: count,
		conf.Download_Part_File_Path_Key:   filePath,
		conf.Download_Part_File_Size_Key:   size,
		conf.Download_Part_Ready_Key:       conf.Download_Ready_Wait,
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
	// 写mq
	msg := &models.TransferMsg{
		TransID:   param.DownloadID,
		FileHash:  hash,
		TmpPath:   filePath,
		FileKey:   fileKey,
		StoreType: conf.Store_Type_COS,
		Task:      conf.Download_Mod,
	}
	err = TransferProduceMsg(msg)
	if err != nil {
		return nil, errors.Wrap(err, "[InitDownload] send msg to MQ error: ")
	}
	// 未下载过或者redis中key已过期
	// redis过期的情况在GetTransList接口中已经处理
	// 即用户已进入trans页面就会将db中过期的记录status改为失败
	// 调用此接口时，不论之前是nil还是fail都应更改状态为process
	// 创建trans记录
	trans := &models.Trans{
		Uuid:           downloadID,
		User_Uuid:      param.User_Uuid,
		User_File_Uuid: param.User_File_Uuid,
		Parent_Uuid:    param.Parent_Uuid,
		File_Key:       fileKey,
		Hash:           hash,
		Local_Path:     param.LocalPath,
		Size:           user_file.Size,
		Name:           user_file.Name,
		Ext:            user_file.Ext,
		Status:         conf.Trans_Process,
		Isdown:         conf.Download_Mod,
	}
	err = client.GetDBClient().CreateTrans(trans)
	if err != nil {
		return nil, errors.Wrap(err, "[InitDownload] set trans record error: ")
	}

	// 返回下载签名
	res.ChunkCount = count
	res.ChunkList = chunkList
	res.DownloadID = downloadID
	res.Hash = hash

	return res, nil
}

func CheckDownloadReady(downloadID string) (string, error) {
	key := helper.GenDownloadPartInfoKey(downloadID)
	// 读取redis
	ready, err := client.GetCacheClient().HGet(key, conf.Download_Part_Ready_Key)
	if err != nil {
		return "", errors.Wrap(err, "[CheckDownloadReady] get download info err: ")
	}
	if ready == "" {
		return "", errors.New("[CheckDownloadReady] get download info err: empty path")
	}
	return ready, nil
}

// 分块下载
func DownloadPart(downloadID string, chunkNum int) ([]byte, error) {
	key := helper.GenDownloadPartInfoKey(downloadID)
	// 读取redis
	path, err := client.GetCacheClient().HGet(key, conf.Download_Part_File_Path_Key)
	if err != nil {
		return nil, errors.Wrap(err, "[DownloadPart] get download info err: ")
	}
	if path == "" {
		return nil, errors.New("[DownloadPart] get download info err: empty path")
	}
	sizeStr, err := client.GetCacheClient().HGet(key, conf.Download_Part_File_Size_Key)
	if err != nil {
		return nil, errors.Wrap(err, "[DownloadPart] get download info err: ")
	}
	size, _ := strconv.Atoi(sizeStr)
	// 打开文件
	fileTmp, err := helper.OpenFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "[DownloadPart] get open file err: ")
	}
	defer fileTmp.Close()
	// 更新redis
	err = client.GetCacheClient().HSet(key, strconv.Itoa(chunkNum), strconv.Itoa(chunkNum))
	if err != nil {
		return nil, errors.Wrap(err, "[DownloadPart] update cache error: ")
	}
	// 写入buf
	buf := make([]byte, conf.File_Part_Size_Max)
	offset := (chunkNum - 1) * conf.File_Part_Size_Max
	fileTmp.Seek(int64(offset), 0)
	// 最后一个分片的大小
	if len(buf) > int(size-offset) {
		buf = make([]byte, size-offset)
	}

	fileTmp.Read(buf)
	// 返回
	return buf, nil
}

// 分块下载完成
func CompleteDownloadPart(downloadID string) error {
	// 查看redis中记录是否完整
	infoKey := helper.GenDownloadPartInfoKey(downloadID)
	infoMap, err := client.GetCacheClient().HGetAll(infoKey)
	if err != nil {
		return errors.Wrap(err, "[CompleteDownloadPart] get download info error: ")
	}
	if _, ok := infoMap[conf.Download_Part_Info_CCount_Key]; !ok {
		return errors.Wrap(conf.MapNotHasError, "[CompleteDownloadPart] get chunk count error: ")
	}
	// 忽略错误
	count, _ := strconv.Atoi(infoMap[conf.Download_Part_Info_CCount_Key])
	// 除去info固定的n个，剩下的fields都对应一个已经上传的分片
	// 如果分片不完整，则返回错误
	if (count) != len(infoMap)-conf.Download_Part_Info_Fileds {
		return errors.Wrap(conf.ChunkMissError, "[CompleteUploadPart] unable to complete: ")
	}
	// 删除文件
	filePath, err := client.GetCacheClient().HGet(infoKey, conf.Download_Part_File_Path_Key)
	if err != nil {
		return errors.Wrap(err, "[CompleteDownloadPart] get file info err: ")
	}
	helper.RemoveDir(filePath)
	// 删除rediskey
	client.GetCacheClient().DelBatch(infoKey)
	// 更改trans表记录状态
	err = client.GetDBClient().UpdateTransState(downloadID, conf.Trans_Success)
	if err != nil {
		return err
	}

	return nil
}
