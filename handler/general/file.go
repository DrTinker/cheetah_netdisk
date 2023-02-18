package general

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// TODO 抽象为service层
// 服务端上传文件，flag为秒传标识
func UploadObjectServer(param *models.UploadObjectParams, fd io.Reader, flag bool) error {
	fileKey := param.FileKey
	user_uuid := param.User_Uuid
	hash := param.Hash
	size := param.Size
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
			return errors.Wrap(err, "[UploadObject] get parent id error: ")
		}
		parentId = ids[user_file_uuid_parent]
	}

	// 从文件KEY中获取文件名称
	name, ext, err := helper.SplitFilePath(fileKey)
	if err != nil {
		return errors.Wrap(err, "[UploadObject] split file key error: ")
	}

	// 拼装结构体
	fileDB := &models.File{
		Uuid: file_uuid,
		Name: name,
		Ext:  ext,
		Path: fileKey,
		Hash: hash,
		Link: 1,
		Size: size,
	}
	userFileDB := &models.UserFile{
		Uuid:      user_file_uuid,
		User_Uuid: user_uuid,
		Parent_Id: parentId,
		File_Uuid: file_uuid,
		Name:      name,
		Ext:       ext,
	}
	// 如果是秒传，只在user_file中插入记录
	if flag {
		err = client.GetDBClient().CreateUserFile(userFileDB)
		if err != nil {
			return errors.Wrap(err, "[UploadObject] store upload record error: ")
		}
		return nil
	}
	// 非秒传
	// 上传COS
	err = client.GetCOSClient().UploadStream(fileKey, fd)
	if err != nil {
		return errors.Wrap(err, "[UploadObject] upload cos error: ")
	}
	// 插入上传记录
	err = client.GetDBClient().CreateUploadRecord(fileDB, userFileDB)
	if err != nil {
		return errors.Wrap(err, "[UploadObject] store upload record error: ")
	}
	return nil
}

func UploadObjectClient(param *models.UploadObjectParams, flag bool) error {
	fileKey := param.FileKey
	user_uuid := param.User_Uuid
	hash := param.Hash
	size := param.Size
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
			return errors.Wrap(err, "[UploadObject] get parent id error: ")
		}
		parentId = ids[user_file_uuid_parent]
	}

	// 从文件KEY中获取文件名称
	name, ext, err := helper.SplitFilePath(fileKey)
	if err != nil {
		return errors.Wrap(err, "[UploadObject] split file key error: ")
	}

	// 拼装结构体
	fileDB := &models.File{
		Uuid: file_uuid,
		Name: name,
		Ext:  ext,
		Path: fileKey,
		Hash: hash,
		Link: 1,
		Size: size,
	}
	userFileDB := &models.UserFile{
		Uuid:      user_file_uuid,
		User_Uuid: user_uuid,
		Parent_Id: parentId,
		File_Uuid: file_uuid,
		Name:      name,
		Ext:       ext,
	}
	// 如果是秒传，只在user_file中插入记录
	if flag {
		err = client.GetDBClient().CreateUserFile(userFileDB)
		if err != nil {
			return errors.Wrap(err, "[UploadObject] store upload record error: ")
		}
		return nil
	}
	// 插入上传记录
	err = client.GetDBClient().CreateUploadRecord(fileDB, userFileDB)
	if err != nil {
		return errors.Wrap(err, "[UploadObject] store upload record error: ")
	}
	return nil
}

// mod: 0复制文件，1移动文件
func CopyObject(src_uuid, des_parent_uuid string) error {
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
		// TODO 引入mq实现异步传输
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
		// DO NOTHING
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
		if err != nil || len(ids) != 2 {
			return errors.Wrap(err, "[CopyObject] get ids error: ")
		}
		user_file_id := ids[user_file_uuid]
		// 对每个文件执行删除
		fList, err := client.GetDBClient().GetUserFileList(user_file_id)
		if err != nil {
			return errors.Wrap(err, "[CopyObject] get user file info error: ")
		}
		for _, file := range fList {
			if err := deleteHelper(file.Uuid, file.File_Uuid); err != nil {
				return err
			}
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
		if err := client.GetDBClient().DeleteUserFileByUuid(file_uuid); err != nil {
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
