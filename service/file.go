package service

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/models"
	"fmt"

	"github.com/pkg/errors"
)

// 创建文件夹
func Mkdir(folder *models.UserFile, ParentUuid string) error {
	if ParentUuid == "" {
		folder.ParentId = conf.DefaultSystemparent
	} else {
		// 获取父级文件夹ID
		ids, err := client.GetDBClient().GetUserFileIDByUuid([]string{ParentUuid})
		if err != nil || ids == nil {
			return errors.Wrap(err, "[Mkdir] get parent id error: ")
		}
		parentId := ids[ParentUuid]
		folder.ParentId = parentId
	}
	// 插入记录
	err := client.GetDBClient().CreateUserFile(folder)
	if err != nil {
		return errors.Wrap(err, "[Mkdir] create user file record error: ")
	}
	return nil
}

// 用于复制和分享
// UserUuid用于区分同用户复制和通过分享链接保存
func CopyObject(src_uuid, des_ParentUuid, UserUuid string) error {
	// 通过uuid获取id
	uuids := []string{src_uuid, des_ParentUuid}
	ids, err := client.GetDBClient().GetUserFileIDByUuid(uuids)
	if err != nil || len(ids) != 2 {
		return errors.Wrap(err, "[CopyObject] get ids error: ")
	}
	src_id := ids[src_uuid]
	des_id := ids[des_ParentUuid]
	// 判断是文件还是文件夹
	user_file, err := client.GetDBClient().GetUserFileByUuid(src_uuid)
	if err != nil || user_file == nil {
		return errors.Wrap(err, "[CopyObject] get user file info error: ")
	}
	// 修改UserUuid
	user_file.UserUuid = UserUuid
	ext := user_file.Ext
	// 如果是文件夹
	if ext == conf.FolderDefaultExt {
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
			// 更改文件夹下每个文件的归属
			f.UserUuid = UserUuid
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
func MoveObject(src_uuid, des_ParentUuid string) error {
	// 通过uuid获取id
	uuids := []string{src_uuid, des_ParentUuid}
	ids, err := client.GetDBClient().GetUserFileIDByUuid(uuids)
	if err != nil || len(ids) != 2 {
		return errors.Wrap(err, "[MoveObject] get ids error: ")
	}

	// 移动
	if err := client.GetDBClient().UpdateUserFileParent(ids[src_uuid], ids[des_ParentUuid]); err != nil {
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
	user_files, err := client.GetDBClient().GetUserFileList(src_file.ParentId)
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
func DeleteObject(UserFileUuid string) error {
	// 获取FileUuid
	user_file, err := client.GetDBClient().GetUserFileByUuid(UserFileUuid)
	if user_file == nil || err != nil {
		return errors.Wrap(err, "[DeleteObject] get user file error: ")
	}
	FileUuid := user_file.FileUuid
	// 判断是文件还是文件夹
	ext := user_file.Ext
	if ext == conf.FolderDefaultExt {
		// 是文件夹
		// 通过uuid获取id
		uuids := []string{UserFileUuid}
		ids, err := client.GetDBClient().GetUserFileIDByUuid(uuids)
		if err != nil {
			return errors.Wrap(err, "[DeleteObject] get ids error: ")
		}
		user_file_id := ids[UserFileUuid]
		// 对每个文件执行删除
		fList, err := client.GetDBClient().GetUserFileList(user_file_id)
		if err != nil {
			return errors.Wrap(err, "[DeleteObject] get user file info error: ")
		}
		for _, file := range fList {
			// 文件夹处理
			if file.Ext == conf.FolderDefaultExt {
				err = client.GetDBClient().DeleteUserFileByUuid(file.Uuid, file.FileUuid)
				if err != nil {
					return errors.Wrap(err, "[DeleteObject] delete user file error: ")
				}
				continue
			}
			// 文件处理
			if err := deleteHelper(file.Uuid, file.FileUuid); err != nil {
				return err
			}
		}
		// 删除user_file中的文件夹记录
		err = client.GetDBClient().DeleteUserFileByUuid(UserFileUuid, FileUuid)
		if err != nil {
			return errors.Wrap(err, "[DeleteObject] delete user file error: ")
		}
	} else {
		// 是文件
		if err := deleteHelper(UserFileUuid, FileUuid); err != nil {
			return err
		}
	}
	return nil
}

func deleteHelper(UserFileUuid, FileUuid string) error {
	// 查看file引用数看是否真正删除
	file, err := client.GetDBClient().GetFileByUuid(FileUuid)
	if err != nil {
		return errors.Wrap(err, "[DeleteObject] get file info error: ")
	}
	// 不为0只删除user_file记录和修改引用数
	if file.Link-1 != 0 {
		if err := client.GetDBClient().DeleteUserFileByUuid(UserFileUuid, file.Uuid); err != nil {
			return errors.Wrap(err, "[DeleteObject] update link error: ")
		}
		return nil
	}
	// 为0同步删除COS和数据库记录
	if err := client.GetCOSClient().Delete(file.FileKey); err != nil {
		return errors.Wrap(err, "[DeleteObject] delete cos error: ")
	}
	if err := client.GetDBClient().DeleteUploadRecord(FileUuid, UserFileUuid); err != nil {
		return errors.Wrap(err, "[DeleteObject] delete db record error: ")
	}
	// 删除分享表数据 忽略错误
	client.GetDBClient().DeleteShareByUserFileUuid(UserFileUuid)
	return nil
}
