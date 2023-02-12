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

func CopyFile(src, des string) error {
	return nil
}
