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

// 文件上传
// 服务端上传文件，flag为秒传标识
// 文件仅在服务端磁盘缓存，缓存后直接向前端返回结果, 通过消息队列异步上传cos
func UploadFileByStream(param *models.TransObjectParams, data []byte) error {
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
	// 写mq
	msg := &models.TransferMsg{
		UploadID:  param.UploadID,
		FileHash:  hash,
		Src:       cfg.TmpPath + "/" + filename,
		Des:       fileKey,
		StoreType: conf.Store_Type_COS,
	}
	err = UploadProduceMsg(msg)
	if err != nil {
		return errors.Wrap(err, "[UploadFileByStream] send msg to MQ error: ")
	}
	// 插入上传记录
	err = client.GetDBClient().CreateUploadRecord(fileDB, userFileDB)
	if err != nil {
		return errors.Wrap(err, "[UploadFileByStream] store upload record error: ")
	}
	// 生成uploadID，并写trans
	uploadID := helper.GenUploadID(user_uuid, hash)
	trans := &models.Trans{
		Uuid:           uploadID,
		User_Uuid:      user_uuid,
		User_File_Uuid: user_file_uuid,
		File_Uuid:      file_uuid,
		File_Key:       fileKey,
		Hash:           hash,
		Size:           size,
		Name:           name,
		Ext:            ext,
		Status:         conf.Trans_Success,
		Isdown:         0,
		Parent_Uuid:    param.Parent,
	}
	// 容忍插入失败
	client.GetDBClient().CreateTrans(trans)
	return nil
}

// 用于分片上传合并文件后上传COS
func UploadFileByPath(param *models.TransObjectParams, path string) error {
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
	}
	// 写mq
	data := &models.TransferMsg{
		UploadID:  param.UploadID,
		FileHash:  hash,
		Src:       path,
		Des:       fileKey,
		StoreType: conf.Store_Type_COS,
	}
	err := UploadProduceMsg(data)
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
	key := helper.GenTransPartInfoKey(uploadID)
	err = client.GetCacheClient().HSet(key, strconv.Itoa(chunkNum), strconv.Itoa(chunkNum))
	if err != nil {
		return errors.Wrap(err, "[UploadPart] update cache error: ")
	}

	return nil
}
