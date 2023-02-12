package general

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"io"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func UploadObject(param *models.UploadObjectParams, fd io.Reader) error {
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
		ids, err := client.GetDBClient().GetFileIDByUuid([]string{user_file_uuid_parent})
		if err != nil || ids == nil {
			return errors.Wrap(err, "[UploadObject] get parent id error: ")
		}
		parentId = ids[0]
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
	// 插入上传记录
	err = client.GetDBClient().CreateUploadRecord(fileDB, userFileDB)
	if err != nil {
		return errors.Wrap(err, "[UploadObject] store upload record error: ")
	}
	// 上传COS
	err = client.GetCOSClient().UploadStream(fileKey, fd)
	if err != nil {
		// 上传失败，回滚数据库
		err = client.GetDBClient().DeleteUploadRecord(file_uuid, user_file_uuid)
		if err != nil {
			// 回滚失败记录日志
			logrus.Warn("[UploadObject] undo upload record failed, file: ", file_uuid, " user_file: ", user_file_uuid)
		}
		return errors.Wrap(err, "[UploadObject] upload cos error: ")
	}
	return nil
}

// mod: 0复制文件，1移动文件
func AlterObject(src_uuid, des_parent_uuid string, mod int) error {
	// 通过uuid获取id
	uuids := []string{src_uuid, des_parent_uuid}
	ids, err := client.GetDBClient().GetFileIDByUuid(uuids)
	if err != nil || len(ids) != 2 {
		return errors.Wrap(err, "[MoveObject] get ids error: ")
	}
	switch mod {
	case 0:
		// 复制
		if err := client.GetDBClient().CopyUserFile(ids[0], ids[1]); err != nil {
			return errors.Wrap(err, "[MoveObject] copy error: ")
		}
	case 1:
		// 移动
		if err := client.GetDBClient().UpdateUserFileParent(ids[0], ids[1]); err != nil {
			return errors.Wrap(err, "[MoveObject] move error: ")
		}
	default:
		return conf.ParamError
	}
	return nil
}
