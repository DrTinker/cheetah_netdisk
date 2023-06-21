package service

import (
	"NetDesk/client"
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"
)

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
		return true, errors.Wrap(err, "[UploadFileByStream] store upload record error: ")
	}
	return true, nil
}

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
		Path:       fileKey,
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
	path := fmt.Sprintf("%s/%s.%s", cfg.TmpPath, hash, ext)
	err = helper.WriteFile(path, data)
	if err != nil {
		return errors.Wrap(err, "[UploadFileByStream] store file to local error: ")
	}
	// 写mq
	msg := &models.TransferMsg{
		FileHash:  hash,
		Src:       path,
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
		Path:       fileKey,
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

// 创建文件夹
func Mkdir(folder *models.UserFile, parent_uuid string) error {
	if parent_uuid == "" {
		folder.Parent_Id = conf.Default_System_parent
	} else {
		// 获取父级文件夹ID
		ids, err := client.GetDBClient().GetUserFileIDByUuid([]string{parent_uuid})
		if err != nil || ids == nil {
			return errors.Wrap(err, "[Mkdir] get parent id error: ")
		}
		parentId := ids[parent_uuid]
		folder.Parent_Id = parentId
	}
	// 插入记录
	err := client.GetDBClient().CreateUserFile(folder)
	if err != nil {
		return errors.Wrap(err, "[Mkdir] create user file record error: ")
	}
	return nil
}

// 用于复制和分享
// user_uuid用于区分同用户复制和通过分享链接保存
func CopyObject(src_uuid, des_parent_uuid, user_uuid string) error {
	// 通过uuid获取id
	uuids := []string{src_uuid, des_parent_uuid}
	ids, err := client.GetDBClient().GetUserFileIDByUuid(uuids)
	if err != nil || len(ids) != 2 {
		return errors.Wrap(err, "[CopyObject] get ids error: ")
	}
	src_id := ids[src_uuid]
	des_id := ids[des_parent_uuid]
	// 判断是文件还是文件夹
	user_file, err := client.GetDBClient().GetUserFileByUuid(src_uuid)
	if err != nil || user_file == nil {
		return errors.Wrap(err, "[CopyObject] get user file info error: ")
	}
	// 修改user_uuid
	user_file.User_Uuid = user_uuid
	ext := user_file.Ext
	// 如果是文件夹
	if ext == conf.Folder_Default_EXT {
		fList, err := client.GetDBClient().GetUserFileList(src_id)
		if err != nil {
			return errors.Wrap(err, "[CopyObject] get user file info error: ")
		}
		// 复制文件夹
		new_des_id, err := client.GetDBClient().CopyUserFile(user_file, des_id)
		if err != nil {
			return errors.Wrap(err, "[CopyObject] copy folder error: ")
		}
		// 复制文件夹下所有文件
		for _, f := range fList {
			if _, err := client.GetDBClient().CopyUserFile(f, new_des_id); err != nil {
				return errors.Wrap(err, fmt.Sprintf("[CopyObject] copy file: %s, err: ", f.Uuid))
			}
		}
	} else {
		// 如果是文件直接复制
		_, err := client.GetDBClient().CopyUserFile(user_file, des_id)
		if err != nil {
			return errors.Wrap(err, "[CopyObject] copy file error: ")
		}
	}
	return nil
}

// 移动文件
func MoveObject(src_uuid, des_parent_uuid string) error {
	// 通过uuid获取id
	uuids := []string{src_uuid, des_parent_uuid}
	ids, err := client.GetDBClient().GetUserFileIDByUuid(uuids)
	if err != nil || len(ids) != 2 {
		return errors.Wrap(err, "[MoveObject] get ids error: ")
	}

	// 移动
	if err := client.GetDBClient().UpdateUserFileParent(ids[src_uuid], ids[des_parent_uuid]); err != nil {
		return errors.Wrap(err, "[MoveObject] move error: ")
	}

	return nil
}

// 修改文件名称，uuid为user_file
func UpdateObjectName(uuid, name, ext string) error {
	// 获取父级文件夹ID，src为要修改的文件
	src_file, err := client.GetDBClient().GetUserFileByUuid(uuid)
	if err != nil {
		return errors.Wrap(err, "[UpdateObjectName] get parent error: ")
	}
	// 获取子文件列表查看重复
	user_files, err := client.GetDBClient().GetUserFileList(src_file.Parent_Id)
	if err != nil {
		return errors.Wrap(err, "[UpdateObjectName] get file list error: ")
	}
	for _, file := range user_files {
		// 重复则返回
		if file.Name == name {
			return conf.NameRepeatError
		}
	}
	// TODO ext变更时处理
	if src_file.Ext != ext {
		return conf.ExtChangeError
	}
	if err := client.GetDBClient().UpdateUserFileName(name, ext, uuid); err != nil {
		return errors.Wrap(err, "[UpdateObjectName] update name error: ")
	}
	return nil
}

// 删除文件
func DeleteObject(user_file_uuid string) error {
	// 获取file_uuid
	user_file, err := client.GetDBClient().GetUserFileByUuid(user_file_uuid)
	if user_file == nil || err != nil {
		return errors.Wrap(err, "[DeleteObject] get user file error: ")
	}
	file_uuid := user_file.File_Uuid
	// 判断是文件还是文件夹
	ext := user_file.Ext
	if ext == conf.Folder_Default_EXT {
		// 是文件夹
		// 通过uuid获取id
		uuids := []string{user_file_uuid}
		ids, err := client.GetDBClient().GetUserFileIDByUuid(uuids)
		if err != nil {
			return errors.Wrap(err, "[DeleteObject] get ids error: ")
		}
		user_file_id := ids[user_file_uuid]
		// 对每个文件执行删除
		fList, err := client.GetDBClient().GetUserFileList(user_file_id)
		if err != nil {
			return errors.Wrap(err, "[DeleteObject] get user file info error: ")
		}
		for _, file := range fList {
			// 文件夹处理
			if file.Ext == conf.Folder_Default_EXT {
				err = client.GetDBClient().DeleteUserFileByUuid(file.Uuid, file.File_Uuid)
				if err != nil {
					return errors.Wrap(err, "[DeleteObject] delete user file error: ")
				}
				continue
			}
			// 文件处理
			if err := deleteHelper(file.Uuid, file.File_Uuid); err != nil {
				return err
			}
		}
		// 删除user_file中的文件夹记录
		err = client.GetDBClient().DeleteUserFileByUuid(user_file_uuid, file_uuid)
		if err != nil {
			return errors.Wrap(err, "[DeleteObject] delete user file error: ")
		}
	} else {
		// 是文件
		if err := deleteHelper(user_file_uuid, file_uuid); err != nil {
			return err
		}
	}
	return nil
}

func deleteHelper(user_file_uuid, file_uuid string) error {
	// 查看file引用数看是否真正删除
	file, err := client.GetDBClient().GetFileByUuid(file_uuid)
	if err != nil {
		return errors.Wrap(err, "[DeleteObject] get file info error: ")
	}
	// 不为0只删除user_file记录和修改引用数
	if file.Link-1 != 0 {
		if err := client.GetDBClient().DeleteUserFileByUuid(user_file_uuid, file.Uuid); err != nil {
			return errors.Wrap(err, "[DeleteObject] update link error: ")
		}
		return nil
	}
	// 为0同步删除COS和数据库记录
	if err := client.GetCOSClient().Delete(file.Path); err != nil {
		return errors.Wrap(err, "[DeleteObject] delete cos error: ")
	}
	if err := client.GetDBClient().DeleteUploadRecord(file_uuid, user_file_uuid); err != nil {
		return errors.Wrap(err, "[DeleteObject] delete db record error: ")
	}
	return nil
}

// 分块上传
// 初始化分块上传，返回UploadID并写入redis
func InitUploadPart(param *models.UploadObjectParams) (*models.UploadPartResult, error) {
	user_uuid := param.User_Uuid
	hash := param.Hash
	size := param.Size

	res := &models.UploadPartResult{}
	// 生成id
	uploadID := helper.GenUploadID(user_uuid, hash)
	key := helper.GenUploadPartInfoKey(uploadID)
	count := size/conf.File_Part_Size_Max + 1
	// 尝试获取分片信息，如果存在则说明之前上传过，触发断点续传逻辑
	tmpInfo, err := client.GetCacheClient().HGetAll(key)
	var chunkList []int
	if tmpInfo != nil && err == nil {
		// 记录已经上传的分片
		for k, _ := range tmpInfo {
			if i, err := strconv.Atoi(k); err == nil {
				chunkList = append(chunkList, i)
			}
		}
		res.ChunkCount = count
		res.ChunkList = chunkList
		res.UploadID = tmpInfo[conf.Upload_Part_Info_ID_Key]
		return res, nil
	}
	// 未上传过或者redis中key已过期
	// 生成分块上传信息
	fileInfo, err := json.Marshal(param)
	if err != nil {
		return nil, errors.Wrap(err, "[InitUploadPart] parse file info error: ")
	}
	info := map[string]interface{}{
		conf.Upload_Part_Info_Hash_Key:   hash,
		conf.Upload_Part_Info_Size_Key:   size,
		conf.Upload_Part_Info_ID_Key:     uploadID,
		conf.Upload_Part_Info_CSize_Key:  conf.File_Part_Size_Max,
		conf.Upload_Part_Info_CCount_Key: count,
		conf.Upload_Part_File_Info_Key:   string(fileInfo),
	}
	// 写redis
	err = client.GetCacheClient().HMSet(key, info)
	if err != nil {
		return nil, errors.Wrap(err, "[InitUploadPart] set upload info error: ")
	}
	err = client.GetCacheClient().Expire(key, conf.Upload_Part_Slice_Expire)
	if err != nil {
		return nil, errors.Wrap(err, "[InitUploadPart] set upload info error: ")
	}
	res.ChunkCount = count
	res.ChunkList = chunkList
	res.UploadID = uploadID
	return res, nil
}

// 分块上传
func UploadPart(uploadID string, chunkNum int, data []byte) error {
	// 写入本地
	cfg, err := client.GetConfigClient().GetLocalConfig()
	if err != nil {
		return errors.Wrap(err, "[UploadPart] get upload config error: ")
	}
	path := fmt.Sprintf("%s/%s/%d", cfg.TmpPath, uploadID, chunkNum)
	err = helper.WriteFile(path, data)
	if err != nil {
		return errors.Wrap(err, "[UploadPart] store file error: ")
	}
	// 更新redis
	key := helper.GenUploadPartInfoKey(uploadID)
	err = client.GetCacheClient().HSet(key, strconv.Itoa(chunkNum), 1)
	if err != nil {
		return errors.Wrap(err, "[UploadPart] update cache error: ")
	}

	return nil
}

func CompleteUploadPart(uploadID string) (*models.UploadObjectParams, string, error) {
	// 查看redis中记录是否全部上传
	infoKey := helper.GenUploadPartInfoKey(uploadID)
	infoMap, err := client.GetCacheClient().HGetAll(infoKey)
	if err != nil {
		return nil, "", errors.Wrap(err, "[CompleteUploadPart] get upload info error: ")
	}
	if _, ok := infoMap[conf.Upload_Part_Info_CCount_Key]; ok {
		return nil, "", errors.Wrap(conf.MapNotHasError, "[CompleteUploadPart] get chunk count error: ")
	}
	// 忽略错误
	count, _ := strconv.Atoi(infoMap[conf.Upload_Part_Info_CCount_Key])
	// 除去info固定的6个，剩下的fields都对应一个已经上传的分片
	// 如果分片不完整，则返回错误
	if (count) != len(infoMap)-conf.Uploac_Part_Info_Fileds {
		return nil, "", errors.Wrap(conf.SliceMissError, "[CompleteUploadPart] unable to complete: ")
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
	src := fmt.Sprintf("/%s/%s/", cfg.TmpPath, uploadID)
	des := fmt.Sprintf("%s/%s.%s", cfg.TmpPath, param.Name, param.Ext)
	fd, err := helper.MergeFile(src, des)
	if err != nil {
		return nil, "", errors.Wrap(err, "[CompleteUploadPart] merge file error: ")
	}
	// 检查文件hash合法性
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, "", errors.Wrap(err, "[CompleteUploadPart] check file hash error: ")
	}
	hash := helper.CountMD5("", data, 1)
	if hash != param.Hash {
		return nil, "", conf.InvaildFileHashError
	}
	return param, des, nil
}

// 下载至tmp
func DownloadToTmp(user_file_uuid string) (string, error) {
	// 通过uuid查询文件信息
	fileKey, err := client.GetDBClient().GetFileKeyByUserFileUuid(user_file_uuid)
	if err != nil {
		return "", err
	}
	// 切分fileKey获取hash&ext
	hash, ext, err := helper.SplitFilePath(fileKey)
	if err != nil {
		return "", errors.Wrap(err, "[DownloadFile] parse fileKey err ")
	}
	// 通过fileKey从COS下载文件
	cfg, err := client.GetConfigClient().GetLocalConfig()
	if err != nil {
		return "", errors.Wrap(err, "[DownloadFile] get config err ")
	}
	path := fmt.Sprintf("%s/%s.%s", cfg.TmpPath, hash, ext)
	err = client.GetCOSClient().DownloadLocal(fileKey, path)
	if err != nil {
		return "", err
	}

	return path, nil
}
