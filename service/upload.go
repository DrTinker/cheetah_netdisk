package service

import (
	"NetDesk/client"
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
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
	user_uuid := param.User_Uuid
	hash := param.Hash
	size := param.Size
	name := param.Name
	ext := param.Ext
	file_uuid := param.File_Uuid
	user_file_uuid := param.User_File_Uuid

	// 获取父级文件夹ID
	var parentId int
	if param.Parent == "" {
		parentId = conf.Default_System_parent
	} else {
		user_file_uuid_parent := param.Parent
		ids, err := client.GetDBClient().GetUserFileIDByUuid([]string{user_file_uuid_parent})
		if err != nil || ids == nil {
			return errors.Wrap(err, "[UploadFileByStream] get parent id error: ")
		}
		parentId = ids[user_file_uuid_parent]
	}

	// 拼装结构体
	fileDB := &models.File{
		Uuid:       file_uuid,
		Name:       name,
		Ext:        ext,
		File_Key:   fileKey,
		Hash:       hash,
		Link:       1,
		Store_Type: conf.Store_Type_Tmp,
		Size:       size,
	}
	userFileDB := &models.UserFile{
		Uuid:      user_file_uuid,
		User_Uuid: user_uuid,
		Parent_Id: parentId,
		File_Uuid: file_uuid,
		Name:      name,
		Ext:       ext,
		Size:      size,
		Hash:      hash,
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
	// 写mq
	msg := &models.TransferMsg{
		TransID:   param.UploadID,
		FileHash:  hash,
		TmpPath:   tmpPath,
		FileKey:   fileKey,
		Thumbnail: tnPath,
		TnFileKey: tnFileKey,
		StoreType: conf.Store_Type_COS,
		Task:      conf.Upload_Mod,
	}
	err = TransferProduceMsg(msg)
	if err != nil {
		// 删除缩略图
		helper.DelFile(tnPath)
		return errors.Wrap(err, "[UploadFileByStream] send msg to MQ error: ")
	}
	// 插入上传记录
	err = client.GetDBClient().CreateUploadRecord(fileDB, userFileDB)
	if err != nil {
		return errors.Wrap(err, "[UploadFileByStream] store upload record error: ")
	}
	// 生成uploadID，并写trans
	// uploadID := helper.GenUploadID(user_uuid, hash)
	trans := &models.Trans{
		Uuid:           param.UploadID,
		User_Uuid:      user_uuid,
		User_File_Uuid: user_file_uuid,
		File_Key:       fileKey,
		Remote_Path:    param.RemotePath,
		Hash:           hash,
		Size:           size,
		Name:           name,
		Ext:            ext,
		Status:         conf.Trans_Success,
		Isdown:         conf.Upload_Mod,
		Parent_Uuid:    param.Parent,
	}
	// 容忍插入失败
	client.GetDBClient().CreateTrans(trans)
	return nil
}

// 用于分片上传合并文件后上传COS
func UploadFileByPath(param *models.UploadObjectParams, path string) error {
	fileKey := param.FileKey
	user_uuid := param.User_Uuid
	hash := param.Hash
	size := param.Size
	name := param.Name
	ext := param.Ext
	file_uuid := param.File_Uuid
	user_file_uuid := param.User_File_Uuid

	// 获取父级文件夹ID
	var parentId int
	if param.Parent == "" {
		parentId = conf.Default_System_parent
	} else {
		user_file_uuid_parent := param.Parent
		ids, err := client.GetDBClient().GetUserFileIDByUuid([]string{user_file_uuid_parent})
		if err != nil || ids == nil {
			return errors.Wrap(err, "[UploadFileByPath] get parent id error: ")
		}
		parentId = ids[user_file_uuid_parent]
	}

	// 拼装结构体
	fileDB := &models.File{
		Uuid:       file_uuid,
		Name:       name,
		Ext:        ext,
		File_Key:   fileKey,
		Hash:       hash,
		Link:       1,
		Store_Type: conf.Store_Type_Tmp,
		Size:       size,
	}
	userFileDB := &models.UserFile{
		Uuid:      user_file_uuid,
		User_Uuid: user_uuid,
		Parent_Id: parentId,
		File_Uuid: file_uuid,
		Name:      name,
		Ext:       ext,
		Size:      size,
		Hash:      hash,
	}
	// 生成缩略图
	src := path
	tnPath, tnName := MediaHandler(src, ext)
	tnFileKey := helper.GenThumbnailKey(tnName)
	fileDB.Thumbnail = tnFileKey
	userFileDB.Thumbnail = tnFileKey
	// 写mq
	msg := &models.TransferMsg{
		TransID:   param.UploadID,
		FileHash:  hash,
		TmpPath:   path,
		FileKey:   fileKey,
		Thumbnail: tnPath,
		TnFileKey: tnFileKey,
		StoreType: conf.Store_Type_COS,
		Task:      conf.Upload_Mod,
	}
	err := TransferProduceMsg(msg)
	if err != nil {
		// 删除缩略图
		helper.DelFile(tnPath)
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
	cfg, err := client.GetConfigClient().GetLocalConfig()
	if err != nil {
		return errors.Wrap(err, "[UploadPart] get upload config error: ")
	}
	path := fmt.Sprintf("%s/%s", cfg.TmpPath, uploadID)
	name := strconv.Itoa(chunkNum)
	err = helper.WriteFile(path, name, data)
	if err != nil {
		return errors.Wrap(err, "[UploadPart] store file error: ")
	}
	// 更新redis
	key := helper.GenUploadPartInfoKey(uploadID)
	err = client.GetCacheClient().HSet(key, strconv.Itoa(chunkNum), strconv.Itoa(chunkNum))
	if err != nil {
		return errors.Wrap(err, "[UploadPart] update cache error: ")
	}

	return nil
}

// 查看文件是否存在
func QuickUpload(param *models.UploadObjectParams) (bool, error) {
	// 处理参数
	user_uuid := param.User_Uuid
	hash := param.Hash
	name := param.Name
	ext := param.Ext
	user_file_uuid := param.User_File_Uuid
	// 通过数据库查询文件是否存在
	flag, file_uuid, err := client.GetDBClient().CheckFileExist(hash)
	if err != nil {
		return false, errors.Wrap(err, "[QuickUpload] get db data error: ")
	}
	// 不存在则返回false
	if !flag {
		return false, nil
	}
	// 存在则触发秒传逻辑
	// 检查是否为同一个人上传同一个文件
	user_flag, err := client.GetDBClient().CheckUserFileExist(user_uuid, file_uuid)
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
		parentId = conf.Default_System_parent
	} else {
		user_file_uuid_parent := param.Parent
		ids, err := client.GetDBClient().GetUserFileIDByUuid([]string{user_file_uuid_parent})
		if err != nil || ids == nil {
			return true, errors.Wrap(err, "[QuickUpload] get parent id error: ")
		}
		parentId = ids[user_file_uuid_parent]
	}

	// 拼接结构体
	userFileDB := &models.UserFile{
		Uuid:      user_file_uuid,
		User_Uuid: user_uuid,
		Parent_Id: parentId,
		File_Uuid: file_uuid,
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
		Uuid:           param.UploadID,
		User_Uuid:      param.User_Uuid,
		User_File_Uuid: param.User_File_Uuid,
		File_Key:       param.FileKey,
		Local_Path:     param.LocalPath,
		Remote_Path:    param.RemotePath,
		Parent_Uuid:    param.Parent,
		Hash:           param.Hash,
		Size:           param.Size,
		Name:           param.Name,
		Ext:            param.Ext,
		Status:         conf.Trans_Success,
		Isdown:         conf.Upload_Mod,
	}
	// 容忍插入失败
	client.GetDBClient().CreateTrans(trans)
	return true, nil
}

// 初始化传输，返回上传签名，并更新redis和db的值
func InitUpload(param *models.UploadObjectParams) (*models.InitUploadResult, error) {
	// 接收参数
	user_uuid := param.User_Uuid
	hash := param.Hash
	size := param.Size
	uploadID := param.UploadID
	// 返回参数
	res := &models.InitUploadResult{}
	// 根据文件大小判断用户空间，是否足够上传
	now, total, err := client.GetDBClient().GetUserVolume(user_uuid)
	if err != nil {
		return nil, err
	}
	if now+int64(size) > total {
		return nil, errors.Wrap(conf.VolumeError, "[InitUpload] user volume err: ")
	}
	// 若为首次传输，则根据文件内容生成uploadID
	if uploadID == "" {
		uploadID = helper.GenUploadID(user_uuid, hash)
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
	count := size/conf.File_Part_Size_Max + 1
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
		res.UploadID = tmpInfo[conf.Upload_Part_Info_ID_Key]
		res.Quick = false
		return res, nil
	}
	// 未上传过或者redis中key已过期
	// redis过期的情况在GetTransList接口中已经处理
	// 即用户已进入trans页面就会将db中过期的记录status改为失败
	// 调用此接口时，不论之前是nil还是fail（success的已经在前面秒传的逻辑处理）都应更改状态为process
	// 创建trans记录
	trans := &models.Trans{
		Uuid:           uploadID,
		User_Uuid:      param.User_Uuid,
		User_File_Uuid: param.User_File_Uuid,
		File_Key:       param.FileKey,
		Local_Path:     param.LocalPath,
		Remote_Path:    param.RemotePath,
		Parent_Uuid:    param.Parent,
		Hash:           param.Hash,
		Size:           param.Size,
		Name:           param.Name,
		Ext:            param.Ext,
		Status:         conf.Trans_Process,
		Isdown:         conf.Upload_Mod,
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
		conf.Upload_Part_Info_ID_Key:     uploadID,
		conf.Upload_Part_Info_CSize_Key:  conf.File_Part_Size_Max,
		conf.Upload_Part_Info_CCount_Key: count,
		conf.Upload_Part_File_Info_Key:   string(fileInfo),
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
	if _, ok := infoMap[conf.Upload_Part_Info_CCount_Key]; !ok {
		return nil, "", errors.Wrap(conf.MapNotHasError, "[CompleteUploadPart] get chunk count error: ")
	}
	// 忽略错误
	count, _ := strconv.Atoi(infoMap[conf.Upload_Part_Info_CCount_Key])
	// 除去info固定的n个，剩下的fields都对应一个已经上传的分片
	// 如果分片不完整，则返回错误
	if (count) != len(infoMap)-conf.Upload_Part_Info_Fileds {
		return nil, "", errors.Wrap(conf.ChunkMissError, "[CompleteUploadPart] unable to complete: ")
	}
	// 从redis中读出init时保存的文件信息
	param := &models.UploadObjectParams{}
	paramStr := infoMap[conf.Upload_Part_File_Info_Key]
	err = json.Unmarshal([]byte(paramStr), param)
	if err != nil {
		return nil, "", errors.Wrap(err, "[CompleteUploadPart] parse file info error: ")
	}
	// 分片完整则，合并文件
	cfg, err := client.GetConfigClient().GetLocalConfig()
	if err != nil {
		return nil, "", errors.Wrap(err, "[CompleteUploadPart] get loacl config error: ")
	}
	src := fmt.Sprintf("%s/%s/", cfg.TmpPath, uploadID)
	des := fmt.Sprintf("%s/%s.%s", cfg.TmpPath, param.Hash, param.Ext)
	_, err = helper.MergeFile(src, des)
	desFlag, _ := helper.PathExists(des)
	// 如果分片文件夹不存在 且 目标文件也不存在
	if err != nil && !desFlag {
		return nil, "", errors.Wrap(err, "[CompleteUploadPart] merge file error: ")
	}
	// 删除分片文件夹
	err = helper.RemoveDir(src[:len(src)-1])
	if err != nil {
		logrus.Warn("[CompleteUploadPart] remove src err: ", err)
	}
	// 检查文件hash合法性
	hash := helper.CountMD5(des, nil, 0)
	if hash != param.Hash {
		// 直接改为失败
		client.GetDBClient().UpdateTransState(uploadID, conf.Trans_Fail)
		// 删除合并后的文件
		helper.DelFile(des)
		return nil, "", conf.InvaildFileHashError
	}
	// 删除redis key
	client.GetCacheClient().DelBatch(infoKey)

	return param, des, nil
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
	infoStr, err := client.GetCacheClient().HGet(infoKey, conf.Upload_Part_File_Info_Key)
	if err != nil {
		return errors.Wrap(err, "[CancelUpload] get upload info err ")
	}
	params := &models.UploadObjectParams{}
	json.Unmarshal([]byte(infoStr), params)
	hash := params.Hash
	if hash == "" {
		return errors.Wrapf(conf.TransBrokenError, "[CancelUpload] upload: %s is broken ", uploadID)
	}
	// 生成分片路径
	cfg, err := client.GetConfigClient().GetLocalConfig()
	if err != nil {
		return errors.Wrap(err, "[CancelUpload] get loacl config error: ")
	}
	src := fmt.Sprintf("%s/%s/", cfg.TmpPath, uploadID)
	flag, _ := helper.PathExists(src)
	// 存在则删除分片
	if flag {
		helper.RemoveDir(src)
	}
	des := fmt.Sprintf("%s/%s.%s", cfg.TmpPath, params.Hash, params.Ext)
	helper.DelFile(des)
	// 删除trans表记录
	err = client.GetDBClient().DelTransByUuid(uploadID)
	if err != nil {
		return errors.Wrap(err, "[CancelUpload] delete trans error: ")
	}
	// 删除redis缓存
	client.GetCacheClient().DelBatch(infoKey)
	return nil
}
