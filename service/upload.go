package service

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// 文件上传
// 服务端上传文件，flag为秒传标识
// 文件仅在服务端磁盘缓存，缓存后直接向前端返回结果, 通过消息队列异步上传cos
func UploadFileByStream(param *models.UploadObjectParams, data []byte) error {
	fileKey := param.FileKey
	UserUuid := param.UserUuid
	hash := param.Hash
	size := param.Size
	name := param.Name
	ext := param.Ext
	FileUuid := param.FileUuid
	UserFileUuid := param.UserFileUuid

	// 获取父级文件夹ID
	var parentId int
	if param.Parent == "" {
		parentId = conf.DefaultSystemparent
	} else {
		UserFileUuid_parent := param.Parent
		ids, err := client.GetDBClient().GetUserFileIDByUuid([]string{UserFileUuid_parent})
		if err != nil || ids == nil {
			return errors.Wrap(err, "[UploadFileByStream] get parent id error: ")
		}
		parentId = ids[UserFileUuid_parent]
	}

	// 拼装结构体
	fileDB := &models.File{
		Uuid:      FileUuid,
		Name:      name,
		Ext:       ext,
		FileKey:   fileKey,
		Hash:      hash,
		Link:      1,
		StoreType: conf.StoreTypeLOS,
		Size:      size,
	}
	userFileDB := &models.UserFile{
		Uuid:     UserFileUuid,
		UserUuid: UserUuid,
		ParentId: parentId,
		FileUuid: FileUuid,
		Name:     name,
		Ext:      ext,
		Size:     size,
		Hash:     hash,
	}
	// 非秒传
	// 先保存本地，再写消息进入消息队列
	cfg, err := client.GetConfigClient().GetLocalConfig()
	if err != nil {
		return errors.Wrap(err, "[UploadFileByStream] get loacl config error: ")
	}
	filename := fmt.Sprintf("%s.%s", hash, ext)
	err = helper.WriteFile(cfg.TmpPath, filename, data)
	if err != nil {
		return errors.Wrap(err, "[UploadFileByStream] store file to local error: ")
	}

	// 生成缩略图
	tmpPath := fmt.Sprintf("%s/%s", cfg.TmpPath, filename)
	tnPath, tnName := MediaHandler(tmpPath, ext)
	tnFileKey := helper.GenThumbnailKey(tnName)
	fileDB.Thumbnail = tnFileKey
	userFileDB.Thumbnail = tnFileKey

	// 删除本地文件
	defer func() {
		helper.DelFile(tmpPath)
		helper.DelFile(tnPath)
	}()

	// 上传私有云
	err = client.GetLOSClient().PutObject(data, fileKey)
	if err != nil {
		return errors.Wrap(err, "[UploadFileByStream] upload los error: ")
	}
	// 缩略图存在才上传
	if tnFileKey != "" && tnPath != "" {
		err = client.GetLOSClient().FPutObject(tnFileKey, tnPath, true)
		if err != nil {
			return errors.Wrap(err, "[UploadFileByStream] upload los error: ")
		}
	}

	// 写mq
	msg := &models.TransferMsg{
		TransID:   param.UploadID,
		FileHash:  hash,
		FileName:  filename,
		TnName:    tnName,
		TmpPath:   fileKey,
		FileKey:   fileKey,
		Thumbnail: tnFileKey,
		TnFileKey: tnFileKey,
		StoreType: conf.StoreTypeCOS,
		Task:      conf.UploadMod,
	}
	err = TransferProduceMsg(msg)
	if err != nil {
		return errors.Wrap(err, "[UploadFileByStream] send msg to MQ error: ")
	}
	// 插入上传记录
	err = client.GetDBClient().CreateUploadRecord(fileDB, userFileDB)
	if err != nil {
		return errors.Wrap(err, "[UploadFileByStream] store upload record error: ")
	}
	// 生成uploadID，并写trans
	// uploadID := helper.GenUploadID(UserUuid, hash)
	trans := &models.Trans{
		Uuid:         param.UploadID,
		UserUuid:     UserUuid,
		UserFileUuid: UserFileUuid,
		FileKey:      fileKey,
		RemotePath:   param.RemotePath,
		Hash:         hash,
		Size:         size,
		Name:         name,
		Ext:          ext,
		Status:       conf.TransSuccess,
		Isdown:       conf.UploadMod,
		ParentUuid:   param.Parent,
	}
	// 容忍插入失败
	client.GetDBClient().CreateTrans(trans)
	return nil
}

// 用于分片上传合并文件后上传COS，在service.CompleteUploadPart后调用
// 文件已在私有云存储中，且service.CompleteUploadPart下载合并后文件到本地
// path: 合并后文件在本地磁盘路径
func UploadFileByPath(param *models.UploadObjectParams, path string) error {
	fileKey := param.FileKey
	UserUuid := param.UserUuid
	hash := param.Hash
	size := param.Size
	name := param.Name
	ext := param.Ext
	FileUuid := param.FileUuid
	UserFileUuid := param.UserFileUuid
	// 删除本地磁盘文件
	defer helper.DelFile(path)

	// 获取父级文件夹ID
	var parentId int
	if param.Parent == "" {
		parentId = conf.DefaultSystemparent
	} else {
		UserFileUuid_parent := param.Parent
		ids, err := client.GetDBClient().GetUserFileIDByUuid([]string{UserFileUuid_parent})
		if err != nil || ids == nil {
			return errors.Wrap(err, "[UploadFileByPath] get parent id error: ")
		}
		parentId = ids[UserFileUuid_parent]
	}

	// 拼装结构体
	fileDB := &models.File{
		Uuid:      FileUuid,
		Name:      name,
		Ext:       ext,
		FileKey:   fileKey,
		Hash:      hash,
		Link:      1,
		StoreType: conf.StoreTypeLOS,
		Size:      size,
	}
	userFileDB := &models.UserFile{
		Uuid:     UserFileUuid,
		UserUuid: UserUuid,
		ParentId: parentId,
		FileUuid: FileUuid,
		Name:     name,
		Ext:      ext,
		Size:     size,
		Hash:     hash,
	}
	// 生成缩略图
	tnPath, tnName := MediaHandler(path, ext)
	// 删除本地磁盘缩略图
	defer helper.DelFile(tnPath)
	// 缩略图存私有云
	tnFileKey := helper.GenThumbnailKey(tnName)
	// 缩略图存在才上传
	if tnFileKey != "" && tnPath != "" {
		err := client.GetLOSClient().FPutObject(tnFileKey, tnPath, true)
		if err != nil {
			errors.Wrap(err, "[UploadFileByPath] los put thumbnail error: ")
		}
	}

	fileDB.Thumbnail = tnFileKey
	userFileDB.Thumbnail = tnFileKey
	// 写mq
	filename := fmt.Sprintf("%s.%s", hash, ext)
	msg := &models.TransferMsg{
		TransID:   param.UploadID,
		FileHash:  hash,
		FileName:  filename,
		TnName:    tnName,
		TmpPath:   fileKey,
		FileKey:   fileKey,
		Thumbnail: tnFileKey,
		TnFileKey: tnFileKey,
		StoreType: conf.StoreTypeCOS,
		Task:      conf.UploadMod,
	}
	err := TransferProduceMsg(msg)
	if err != nil {
		return errors.Wrap(err, "[UploadFileByPath] send msg to MQ error: ")
	}
	// 插入上传记录
	err = client.GetDBClient().CreateUploadRecord(fileDB, userFileDB)
	if err != nil {
		return errors.Wrap(err, "[UploadFileByPath] store upload record error: ")
	}
	return nil
}

// 分块上传
func UploadPart(uploadID string, chunkNum int, data []byte) error {
	// 写入本地
	// cfg, err := client.GetConfigClient().GetLocalConfig()
	// if err != nil {
	// 	return errors.Wrap(err, "[UploadPart] get upload config error: ")
	// }

	// path := fmt.Sprintf("%s/%s", cfg.TmpPath, uploadID)
	// name := strconv.Itoa(chunkNum)
	// err = helper.WriteFile(path, name, data)

	// 获取分片总数，用来填充0
	key := helper.GenUploadPartInfoKey(uploadID)
	chunkCntStr, err := client.GetCacheClient().HGet(key, conf.UploadPartInfoCCountKey)
	chunkCnt, _ := strconv.Atoi(chunkCntStr)
	if err != nil || chunkCnt == 0 {
		return conf.ChunkMissError
	}
	zero := helper.CountDigit(chunkCnt)
	// 分片在私有云中的存储路径
	partKey := fmt.Sprintf("%s/%s/%0*d", conf.TmpPrefix, uploadID, zero, chunkNum)

	// 写入MinIO
	err = client.GetLOSClient().PutObject(data, partKey)
	if err != nil {
		return errors.Wrap(err, "[UploadPart] los put object error: ")
	}
	// 更新redis
	err = client.GetCacheClient().HSet(key, strconv.Itoa(chunkNum), strconv.Itoa(chunkNum))
	if err != nil {
		return errors.Wrap(err, "[UploadPart] update cache error: ")
	}

	return nil
}

// 查看文件是否存在
func QuickUpload(param *models.UploadObjectParams) (bool, error) {
	// 处理参数
	UserUuid := param.UserUuid
	hash := param.Hash
	name := param.Name
	ext := param.Ext
	UserFileUuid := param.UserFileUuid
	// 通过数据库查询文件是否存在
	flag, file, err := client.GetDBClient().CheckFileExist(hash)
	if err != nil {
		return false, errors.Wrap(err, "[QuickUpload] get db data error: ")
	}
	// 不存在则返回false
	if !flag {
		return false, nil
	}
	// 存在则触发秒传逻辑
	// 检查是否为同一个人上传同一个文件
	user_flag, err := client.GetDBClient().CheckUserFileExist(UserUuid, file.Uuid)
	if err != nil {
		return false, errors.Wrap(err, "[QuickUpload] get db data error: ")
	}
	// 已经在该用户空间存在时应该走复制接口
	if user_flag {
		return false, conf.FileExistError
	}
	// 获取父级文件夹ID
	var parentId int
	if param.Parent == "" {
		parentId = conf.DefaultSystemparent
	} else {
		UserFileUuid_parent := param.Parent
		ids, err := client.GetDBClient().GetUserFileIDByUuid([]string{UserFileUuid_parent})
		if err != nil || ids == nil {
			return true, errors.Wrap(err, "[QuickUpload] get parent id error: ")
		}
		parentId = ids[UserFileUuid_parent]
	}

	// 拼接结构体
	userFileDB := &models.UserFile{
		Uuid:      UserFileUuid,
		UserUuid:  UserUuid,
		ParentId:  parentId,
		FileUuid:  file.Uuid,
		Thumbnail: file.Thumbnail,
		Size:      param.Size,
		Name:      name,
		Ext:       ext,
		Hash:      hash,
	}
	// 秒传只写user_file表
	err = client.GetDBClient().CreateQuickUploadRecord(userFileDB, param.Size)
	if err != nil {
		return true, errors.Wrap(err, "[QuickUpload] store upload record error: ")
	}
	// 创建trans记录
	trans := &models.Trans{
		Uuid:         param.UploadID,
		UserUuid:     param.UserUuid,
		UserFileUuid: param.UserFileUuid,
		FileKey:      param.FileKey,
		LocalPath:    param.LocalPath,
		RemotePath:   param.RemotePath,
		ParentUuid:   param.Parent,
		Hash:         param.Hash,
		Size:         param.Size,
		Name:         param.Name,
		Ext:          param.Ext,
		Status:       conf.TransSuccess,
		Isdown:       conf.UploadMod,
	}
	// 容忍插入失败
	client.GetDBClient().CreateTrans(trans)
	return true, nil
}

// 初始化传输，返回上传签名，并更新redis和db的值
func InitUpload(param *models.UploadObjectParams) (*models.InitUploadResult, error) {
	// 接收参数
	UserUuid := param.UserUuid
	hash := param.Hash
	size := param.Size
	uploadID := param.UploadID
	// 返回参数
	res := &models.InitUploadResult{}
	// 根据文件大小判断用户空间，是否足够上传
	now, total, err := client.GetDBClient().GetUserVolume(UserUuid)
	if err != nil {
		return nil, err
	}
	if now+int64(size) > total {
		return nil, errors.Wrap(conf.VolumeError, "[InitUpload] user volume err: ")
	}
	// 若为首次传输，则根据文件内容生成uploadID
	if uploadID == "" {
		uploadID = helper.GenUploadID(UserUuid, hash)
		param.UploadID = uploadID
		// 判断秒传，若为秒传则记录数据库
		quick, err := QuickUpload(param)
		if err == conf.FileExistError {
			return nil, err
		}
		// 秒传直接返回
		if quick {
			res.UploadID = uploadID
			res.Quick = true
			return res, nil
		}
	}

	// 非秒传或者断点续传
	// 尝试获取分片信息，如果存在则说明之前上传过，触发断点续传逻辑
	infoKey := helper.GenUploadPartInfoKey(uploadID)
	count := size/conf.FilePartSizeMax + 1
	tmpInfo, err := client.GetCacheClient().HGetAll(infoKey)
	var chunkList []int
	if len(tmpInfo) != 0 && err == nil {
		// 记录已经上传的分片
		for k, _ := range tmpInfo {
			if i, err := strconv.Atoi(k); err == nil {
				chunkList = append(chunkList, i)
			}
		}
		res.ChunkCount = count
		res.ChunkList = chunkList
		res.UploadID = tmpInfo[conf.UploadPartInfoIDKey]
		res.Quick = false
		return res, nil
	}
	// 未上传过或者redis中key已过期
	// redis过期的情况在GetTransList接口中已经处理
	// 即用户已进入trans页面就会将db中过期的记录status改为失败
	// 调用此接口时，不论之前是nil还是fail（success的已经在前面秒传的逻辑处理）都应更改状态为process
	// 创建trans记录
	trans := &models.Trans{
		Uuid:         uploadID,
		UserUuid:     param.UserUuid,
		UserFileUuid: param.UserFileUuid,
		FileKey:      param.FileKey,
		LocalPath:    param.LocalPath,
		RemotePath:   param.RemotePath,
		ParentUuid:   param.Parent,
		Hash:         param.Hash,
		Size:         param.Size,
		Name:         param.Name,
		Ext:          param.Ext,
		Status:       conf.TransProcess,
		Isdown:       conf.UploadMod,
	}
	err = client.GetDBClient().CreateTrans(trans)
	if err != nil {
		return nil, errors.Wrap(err, "[InitUpload] set trans record error: ")
	}

	// 生成分块上传信息
	fileInfo, err := json.Marshal(param)
	if err != nil {
		return nil, errors.Wrap(err, "[InitUploadPart] parse file info error: ")
	}
	info := map[string]interface{}{
		conf.UploadPartInfoIDKey:     uploadID,
		conf.UploadPartInfoCSizeKey:  conf.FilePartSizeMax,
		conf.UploadPartInfoCCountKey: count,
		conf.UploadPartFileInfoKey:   string(fileInfo),
	}
	// 写redis
	err = client.GetCacheClient().HMSet(infoKey, info)
	if err != nil {
		return nil, errors.Wrap(err, "[InitUpload] set upload info error: ")
	}
	// 设置过期时间
	err = client.GetCacheClient().Expire(infoKey, conf.Trans_Part_Slice_Expire)
	if err != nil {
		return nil, errors.Wrap(err, "[InitUploadPart] set upload info error: ")
	}

	// 返回上传签名
	res.ChunkCount = count
	res.ChunkList = chunkList
	res.UploadID = uploadID

	return res, nil
}

// 上传完成合并分片
func CompleteUploadPart(uploadID string) (*models.UploadObjectParams, string, error) {
	// 查看redis中记录是否全部上传
	infoKey := helper.GenUploadPartInfoKey(uploadID)
	infoMap, err := client.GetCacheClient().HGetAll(infoKey)
	if err != nil {
		return nil, "", errors.Wrap(err, "[CompleteUploadPart] get upload info error: ")
	}
	if _, ok := infoMap[conf.UploadPartInfoCCountKey]; !ok {
		return nil, "", errors.Wrap(conf.MapNotHasError, "[CompleteUploadPart] get chunk count error: ")
	}
	// 忽略错误
	count, _ := strconv.Atoi(infoMap[conf.UploadPartInfoCCountKey])
	// 除去info固定的n个，剩下的fields都对应一个已经上传的分片
	// 如果分片不完整，则返回错误
	if (count) != len(infoMap)-conf.UploadPartInfoFileds {
		return nil, "", errors.Wrap(conf.ChunkMissError, "[CompleteUploadPart] unable to complete: ")
	}
	// 从redis中读出init时保存的文件信息
	param := &models.UploadObjectParams{}
	paramStr := infoMap[conf.UploadPartFileInfoKey]
	err = json.Unmarshal([]byte(paramStr), param)
	if err != nil {
		return nil, "", errors.Wrap(err, "[CompleteUploadPart] parse file info error: ")
	}
	// 分片完整则，合并文件
	// 存本地的弃用
	// cfg, err := client.GetConfigClient().GetLocalConfig()
	// if err != nil {
	// 	return nil, "", errors.Wrap(err, "[CompleteUploadPart] get loacl config error: ")
	// }
	// src := fmt.Sprintf("%s/%s/", cfg.TmpPath, uploadID)
	// des := fmt.Sprintf("%s/%s.%s", cfg.TmpPath, param.Hash, param.Ext)
	// _, err = helper.MergeFile(src, des)
	// desFlag, _ := helper.PathExists(des)
	// // 如果分片文件夹不存在 且 目标文件也不存在
	// if err != nil && !desFlag {
	// 	return nil, "", errors.Wrap(err, "[CompleteUploadPart] merge file error: ")
	// }
	// // 删除分片文件夹
	// err = helper.RemoveDir(src[:len(src)-1])
	// if err != nil {
	// 	logrus.Warn("[CompleteUploadPart] remove src err: ", err)
	// }
	// // 检查文件hash合法性
	// hash := helper.CountMD5(des, nil, 0)
	// if hash != param.Hash {
	// 	// 直接改为失败
	// 	client.GetDBClient().UpdateTransState(uploadID, conf.TransFail)
	// 	// 删除合并后的文件
	// 	helper.DelFile(des)
	// 	return nil, "", conf.InvaildFileHashError
	// }
	src := fmt.Sprintf("%s/%s/", conf.TmpPrefix, uploadID)
	des := fmt.Sprintf("%s/%s.%s", conf.FilePrefix, param.Hash, param.Ext)
	contentType := "application/octet-stream"
	if ct, ok := helper.ExtToContentType[fmt.Sprintf(".%s", param.Ext)]; ok {
		contentType = ct
	}
	// 在私有云中合并 合并成功后清除分片文件
	err = client.GetLOSClient().MergeObjects(src, des, contentType, true)
	if err != nil {
		return nil, "", errors.Wrap(err, "[CompleteUploadPart] los merge file error: ")
	}
	// 从私有云下载合并后的文件到本地磁盘
	cfg, err := client.GetConfigClient().GetLocalConfig()
	if err != nil {
		return nil, "", errors.Wrap(err, "[CompleteUploadPart] get loacl config error: ")
	}
	filePath := fmt.Sprintf("%s/%s.%s", cfg.TmpPath, param.Hash, param.Ext)
	err = client.GetLOSClient().FGetObject(des, filePath)
	if err != nil {
		return nil, "", errors.Wrap(err, "[CompleteUploadPart] los get file error: ")
	}
	// 检查文件hash合法性
	hash := helper.CountMD5(filePath, nil, 0)
	if hash != param.Hash {
		// 删除本地文件
		helper.DelFile(filePath)
		// 直接改为失败
		client.GetDBClient().UpdateTransState(uploadID, conf.TransFail)
		return nil, "", conf.InvaildFileHashError
	}
	// 删除redis key
	client.GetCacheClient().DelBatch(infoKey)

	// TODO 这里存本地和存私有云返回有区别
	return param, filePath, nil
}

func CancelUpload(uploadID string) error {
	// redis查看是否存在记录，若不存在则一定不在进行中
	infoKey := helper.GenUploadPartInfoKey(uploadID)
	num, err := client.GetCacheClient().Exists(infoKey)
	if err != nil {
		return errors.Wrap(err, "[CancelUpload] get upload info err ")
	}
	// 不存在直接说明已经结束或者失败
	if num == 0 {
		return errors.Wrapf(conf.TransFinishError, "[CancelUpload] upload: %s is finished ", uploadID)
	}
	// 存在说明正在进行
	// 读取redis获取hash
	infoStr, err := client.GetCacheClient().HGet(infoKey, conf.UploadPartFileInfoKey)
	if err != nil {
		return errors.Wrap(err, "[CancelUpload] get upload info err ")
	}
	params := &models.UploadObjectParams{}
	json.Unmarshal([]byte(infoStr), params)
	hash := params.Hash
	if hash == "" {
		return errors.Wrapf(conf.TransBrokenError, "[CancelUpload] upload: %s is broken ", uploadID)
	}
	// // 生成分片路径
	// cfg, err := client.GetConfigClient().GetLocalConfig()
	// if err != nil {
	// 	return errors.Wrap(err, "[CancelUpload] get loacl config error: ")
	// }
	// src := fmt.Sprintf("%s/%s/", cfg.TmpPath, uploadID)
	// flag, _ := helper.PathExists(src)
	// // 存在则删除分片
	// if flag {
	// 	helper.RemoveDir(src)
	// }
	// des := fmt.Sprintf("%s/%s.%s", cfg.TmpPath, params.Hash, params.Ext)
	// helper.DelFile(des)
	// 生成分片路径
	src := fmt.Sprintf("%s/%s/", conf.TmpPrefix, uploadID)
	err = client.GetLOSClient().RemoveDir(src)
	if err != nil {
		logrus.Error(err, "[CancelUpload] remove parts error:")
	}
	// 删除合并后的文件 如有
	des := fmt.Sprintf("%s/%s.%s", conf.FilePrefix, params.Hash, params.Ext)
	client.GetLOSClient().RemoveObject(des)
	// 删除trans表记录
	err = client.GetDBClient().DelTransByUuid(uploadID)
	if err != nil {
		return errors.Wrap(err, "[CancelUpload] delete trans error: ")
	}
	// 删除redis缓存
	client.GetCacheClient().DelBatch(infoKey)
	return nil
}
