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
)

// 文件传输相关

// 查看文件是否存在
func QuickUpload(param *models.TransObjectParams) (bool, error) {
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
	user_flag, err := client.GetDBClient().CheckUserFileExist(file_uuid, user_uuid)
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
		File_Uuid:      param.File_Uuid,
		File_Key:       param.FileKey,
		Local_Path:     param.LocalPath,
		Hash:           param.Hash,
		Size:           param.Size,
		Name:           param.Name,
		Ext:            param.Ext,
		Status:         conf.Trans_Success,
	}
	err = client.GetDBClient().CreateTrans(trans)
	if err != nil {
		return true, errors.Wrap(err, "[QuickUpload] record trans error: ")
	}
	return true, nil
}

// 初始化传输，返回上传签名，并更新redis和db的值
func InitUpload(param *models.TransObjectParams) (*models.InitTransResult, error) {
	// 接收参数
	user_uuid := param.User_Uuid
	hash := param.Hash
	size := param.Size
	uploadID := param.UploadID
	// 返回参数
	res := &models.InitTransResult{}
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
		File_Uuid:      param.File_Uuid,
		File_Key:       param.FileKey,
		Local_Path:     param.LocalPath,
		Parent_Uuid:    param.Parent,
		Hash:           param.Hash,
		Size:           param.Size,
		Name:           param.Name,
		Ext:            param.Ext,
		Status:         conf.Trans_Process,
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
	err = client.GetCacheClient().Expire(infoKey, conf.Upload_Part_Slice_Expire)
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
func CompleteUploadPart(uploadID string) (*models.TransObjectParams, string, error) {
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
		return nil, "", errors.Wrap(conf.SliceMissError, "[CompleteUploadPart] unable to complete: ")
	}
	// 从redis中读出init时保存的文件信息
	param := &models.TransObjectParams{}
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
	des := fmt.Sprintf("%s/%s.%s", cfg.TmpPath, param.Name, param.Ext)
	_, err = helper.MergeFile(src, des)
	if err != nil {
		return nil, "", errors.Wrap(err, "[CompleteUploadPart] merge file error: ")
	}
	// 检查文件hash合法性
	hash := helper.CountMD5(des, nil, 0)
	if hash != param.Hash {
		return nil, "", conf.InvaildFileHashError
	}
	return param, des, nil
}
