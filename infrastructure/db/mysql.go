package db

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
)

type DBClientImpl struct {
	DBConn *gorm.DB
}

func NewDBClientImpl(driver, source string) (*DBClientImpl, error) {
	db := &DBClientImpl{}
	conn, _ := gorm.Open(mysql.Open(source), &gorm.Config{NamingStrategy: schema.NamingStrategy{
		SingularTable: true, // 指定单数表名
	}})
	// debug模式
	//conn.LogMode(true)
	// 全局禁用复数表名
	sqlDB, err := conn.DB()
	sqlDB.SetMaxOpenConns(conf.MaxConn)
	sqlDB.SetMaxIdleConns(conf.MaxIdleConn)
	sqlDB.SetConnMaxIdleTime(conf.MaxIdleTime)
	if err != nil {
		return nil, err
	}
	db.DBConn = conn
	return db, nil
}

// 用户
// 创建用户，同时创建用户文件空间根目录
func (d *DBClientImpl) CreateUser(user *models.User, folder *models.UserFile) error {
	err := d.DBConn.Transaction(func(tx *gorm.DB) error {
		err := tx.Table(conf.UserTB).Create(user).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] CreateUser Create user err:")
		}

		err = tx.Table(conf.UserFileTB).Create(folder).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] CreateUser Create folder err:")
		}
		return nil
	})

	return err
}

// 通过ID检索
func (d *DBClientImpl) GetUserByID(id string) (*models.User, error) {
	user := &models.User{}
	// 查询用户信息
	err := d.DBConn.Table(conf.UserTB).Where(conf.UserUuidDB+"=?", id).First(user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(err, "[DBClientImpl] GetUserByID Select err:")
	}

	return user, nil
}

// 通过邮箱检索
func (d *DBClientImpl) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	// 查询用户信息
	err := d.DBConn.Table(conf.UserTB).Where(conf.UserEmailDB+"=?", email).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, conf.DBNotFoundError
		}
		return nil, errors.Wrap(err, "[DBClientImpl] GetUserByEmail Select err:")
	}

	return user, nil
}

func (d *DBClientImpl) GetUserVolume(id string) (now, total int64, err error) {
	user := &models.User{}
	err = d.DBConn.Table(conf.UserTB).Where(conf.UserUuidDB+"=?", id).First(user).Error
	if err != nil {
		return 0, 0, errors.Wrap(err, "[DBClientImpl] GetUserVolume Select err:")
	}
	return user.NowVolume, user.TotalVolume, nil
}

// 更改用户名
func (d DBClientImpl) UpdateUserName(uuid, name string) error {
	err := d.DBConn.Table(conf.UserTB).Where(conf.UserUuidDB+"=?", uuid).
		Update(conf.UserNameDB, name).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] UpdateUserName update user name err: ")
	}
	return nil
}

// 文件
// 检测文件存在性,false表示不存在
func (d *DBClientImpl) CheckFileExist(hash string) (bool, *models.File, error) {
	file := &models.File{}
	err := d.DBConn.Table(conf.FilePoolTB).Where(conf.FileHashDB+"=?", hash).First(file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil
		}
		return false, nil, errors.Wrap(err, "[DBClientImpl] CheckFileExist Select err:")
	}
	return true, file, nil
}

func (d *DBClientImpl) CheckUserFileExist(UserUuid, FileUuid string) (bool, error) {
	user_file := &models.UserFile{}
	err := d.DBConn.Table(conf.UserFileTB).Where(&models.UserFile{FileUuid: FileUuid, UserUuid: UserUuid}).
		First(user_file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, errors.Wrap(err, "[DBClientImpl] CheckUserFileExist Select err:")
	}
	return true, nil
}

// 存储上传文件记录
func (d *DBClientImpl) CreateUploadRecord(file *models.File, userFile *models.UserFile) error {
	err := d.DBConn.Transaction(func(tx *gorm.DB) error {
		// 从这里开始使用 'tx' 而不是 'db'
		// 检查文件大小
		user := &models.User{}
		err := tx.Table(conf.UserTB).Where(conf.UserUuidDB+"=?", userFile.UserUuid).First(user).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] GetUserVolume Select err:")
		}
		cur := user.NowVolume + int64(file.Size)
		if user.TotalVolume < cur+int64(file.Size) {
			return conf.VolumeError
		}
		// user_file表增加记录
		if err := tx.Table(conf.UserFileTB).Create(userFile).Error; err != nil {
			// 返回任何错误都会回滚事务
			return errors.Wrap(err, "[DBClientImpl] CreateUploadRecord Create user file err:")
		}
		// file_pool表增加记录
		if err := tx.Table(conf.FilePoolTB).Create(file).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CreateUploadRecord Create file err:")
		}
		// 更新用户空间大小
		if err := tx.Table(conf.UserTB).Where(conf.UserUuidDB+"=?", userFile.UserUuid).
			Update(conf.UserNowVolumeDB, cur).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CreateUploadRecord Update user err:")
		}
		// 返回 nil 提交事务
		return nil
	})
	return err
}

// 秒传文件记录
func (d *DBClientImpl) CreateQuickUploadRecord(userFile *models.UserFile, size int) error {
	err := d.DBConn.Transaction(func(tx *gorm.DB) error {
		// 从这里开始使用 'tx' 而不是 'db'
		// user_file表增加记录
		if err := tx.Table(conf.UserFileTB).Create(userFile).Error; err != nil {
			// 返回任何错误都会回滚事务
			return errors.Wrap(err, "[DBClientImpl] CreateQuickUploadRecord Create user file err:")
		}
		// 更新用户空间大小
		if err := tx.Table(conf.UserTB).Where(conf.UserUuidDB+"=?", userFile.UserUuid).
			Update(conf.UserNowVolumeDB, gorm.Expr(conf.UserNowVolumeDB+"+?", size)).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CreateQuickUploadRecord Update user err:")
		}
		// 修改file_pool中文件引用数
		if err := tx.Table(conf.FilePoolTB).Where(conf.FileUuidDB+"=?", userFile.FileUuid).
			Update(conf.FileLinkDB, gorm.Expr(conf.FileLinkDB+"+?", 1)).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CreateQuickUploadRecord Update link err:")
		}
		// 返回 nil 提交事务
		return nil
	})
	return err
}

// 删除上传记录，用于回滚
func (d *DBClientImpl) DeleteUploadRecord(FileUuid, UserFileUuid string) error {
	err := d.DBConn.Transaction(func(tx *gorm.DB) error {
		userFile := &models.UserFile{}
		// 先查询user_file，mysql不支持returning
		err := tx.Table(conf.UserFileTB).
			Where(conf.UserFileUuidDB+"=?", UserFileUuid).Find(userFile).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] DeleteUserFileByUuid get user file err:")
		}
		// user_file表删除记录，软删除, 返回被删除的列
		if err := tx.Table(conf.UserFileTB).
			Where(conf.UserFileUuidDB+"=?", UserFileUuid).Delete(userFile).Error; err != nil {
			// 返回任何错误都会回滚事务
			return errors.Wrap(err, "[DBClientImpl] DeleteUploadRecord delete user file err:")
		}
		file := &models.File{}
		// file_pool表删除记录，真删除，与cos保持同步，当cos中文件被删除后用户空间无法恢复
		if err := tx.Table(conf.FilePoolTB).Where(conf.FileUuidDB+"=?", FileUuid).Unscoped().Delete(file).Error; err != nil {
			// 返回任何错误都会回滚事务
			return errors.Wrap(err, "[DBClientImpl] DeleteUploadRecord delete user file err:")
		}
		// 更新用户空间大小
		if err := tx.Table(conf.UserTB).Where(conf.UserUuidDB+"=?", userFile.UserUuid).
			Update(conf.UserNowVolumeDB, gorm.Expr(conf.UserNowVolumeDB+"-?", userFile.Size)).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CreateQuickUploadRecord Update user err:")
		}
		return nil
	})
	return err
}

// 获取父级文件
func (d *DBClientImpl) GetUserFileParent(uuid string) (file *models.UserFile, err error) {
	file = &models.UserFile{}
	err = d.DBConn.Table(conf.UserFileTB).Where(conf.UserFileUuidDB+"=?", uuid).Find(file).Error
	if err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetUserFileParent err:")
	}
	return file, nil
}

// 获取文件列表
func (d *DBClientImpl) GetUserFileList(ParentId int) (files []*models.UserFile, err error) {
	// find无需初始化，且应传入数组指针
	err = d.DBConn.Table(conf.UserFileTB).Where(conf.UserFileParentDB+"=?", ParentId).Find(&files).Error
	if files == nil || err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetFileList err:")
	}
	return
}

// 分页获取文件列表
func (d *DBClientImpl) GetUserFileListPage(ParentId int, cur, pageSize int, ext string) (files []*models.UserFile, err error) {

	var findMode *models.UserFile
	if ext == "" {
		findMode = &models.UserFile{ParentId: ParentId}
	} else {
		findMode = &models.UserFile{ParentId: ParentId, Ext: ext}
	}

	// limit + offset实现分页查询
	err = d.DBConn.Table(conf.UserFileTB).Where(findMode).
		Limit(pageSize).Offset((cur - 1) * pageSize).Find(&files).Error
	if files == nil || err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetUserFileListPage err:")
	}
	return
}

// 通过文件uuid获取id
func (d *DBClientImpl) GetUserFileIDByUuid(uuids []string) (ids map[string]int, err error) {
	var files []models.UserFile
	ids = make(map[string]int, len(uuids))
	// 查出来file是乱序的
	err = d.DBConn.Table(conf.UserFileTB).Where(conf.FileUuidDB+" in (?)", uuids).
		Select(conf.UserFileIdDB, conf.UserFileUuidDB).Find(&files).Error
	if len(files) == 0 || err != nil {
		if err == nil {
			err = conf.DBNotFoundError
		}
		return nil, errors.Wrap(err, "[DBClientImpl] GetUserFileIDByUuid err:")
	}
	for _, file := range files {
		uuid := file.Uuid
		ids[uuid] = int(file.ID)
	}
	return ids, nil
}

// 通过COS文件唯一KEY获取用户文件信息
func (d *DBClientImpl) GetUserFileByPath(path string) (user_file *models.UserFile, err error) {
	user_file = &models.UserFile{}
	// 拼接sql
	ft := conf.FilePoolTB
	fid := conf.FileUuidDB
	uft := conf.UserFileTB
	ufid := conf.UserFilePoolUuidDB
	// select * from "user_file" inner join "file_pool" on ("user_file".FileUuid = "file_pool".uuid) where "file_pool".path = path
	err = d.DBConn.Table(conf.UserFileTB).Joins(fmt.Sprintf("inner join %s on %s.%s = %s.%s", ft, ft, fid, uft, ufid)).
		Where(fmt.Sprintf("%s.%s=?", ft, conf.FileFileKeyDB), path).First(user_file).Error
	if err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetUserFileByPath err:")
	}
	return user_file, nil
}

// 获取单个用户文件信息，用于复制和移动
func (d *DBClientImpl) GetUserFileByUuid(uuid string) (file *models.UserFile, err error) {
	file = &models.UserFile{}
	err = d.DBConn.Table(conf.UserFileTB).Where(conf.FileUuidDB+"=?", uuid).First(file).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(err, "[DBClientImpl] GetUserFileByUuid err:")
	}
	return file, nil
}

// 获取批量用户文件信息，用于批量复制和移动
func (d *DBClientImpl) GetUserFileBatch(uuids []string) (files []*models.UserFile, err error) {
	files = make([]*models.UserFile, len(uuids))
	err = d.DBConn.Table(conf.UserFileTB).Where(conf.FileUuidDB+" in (?)", uuids).Find(&files).Error
	if err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetUserFileBatch err:")
	}
	return files, nil
}

func (d *DBClientImpl) GetUserByFileUuid(FileUuid string) (UserUuid string, err error) {
	user_file := &models.UserFile{}
	err = d.DBConn.Table(conf.UserFileTB).Where(conf.UserFilePoolUuidDB+"=?", FileUuid).
		Select(conf.UserFileUserIdDB).Find(user_file).Error
	if err != nil {
		return "", errors.Wrap(err, "[DBClientImpl] GetUserByFileUuid err:")
	}
	return user_file.UserUuid, nil
}

// 通过UserFileUuid查询对应的FileUuid
func (d *DBClientImpl) GetFileUuidByUserFileUuid(UserFileUuid string) (FileUuid string, err error) {
	user_file := &models.UserFile{}
	err = d.DBConn.Table(conf.UserFileTB).Where(conf.UserFileUuidDB+"=?", UserFileUuid).
		Select(conf.UserFilePoolUuidDB).Find(user_file).Error
	if err != nil {
		return "", errors.Wrap(err, "[DBClientImpl] GetFileUuidByUserFileUuid err:")
	}
	return user_file.FileUuid, nil
}

// 在用户文件空间复制
func (d *DBClientImpl) CopyUserFile(src_file *models.UserFile, des_ParentId int) (int, error) {
	// 生成新id, uuid和parentId
	copy_file := &models.UserFile{
		Uuid:      helper.GenUserFid(src_file.UserUuid, src_file.Name+"_copy"),
		ParentId:  des_ParentId,
		Name:      src_file.Name,
		Ext:       src_file.Ext,
		Size:      src_file.Size,
		Thumbnail: src_file.Thumbnail,
		Hash:      src_file.Hash,
		UserUuid:  src_file.UserUuid,
		FileUuid:  src_file.FileUuid,
	}
	// 复制文件
	if err := d.DBConn.Transaction(func(tx *gorm.DB) error {
		// 复制user_file记录
		if err := tx.Table(conf.UserFileTB).Create(copy_file).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CopyUserFile create copy err:")
		}
		// 增加file_pool中link数
		FileUuid := src_file.FileUuid
		if err := tx.Table(conf.FilePoolTB).Where(conf.FileUuidDB+"=?", FileUuid).
			Update(conf.FileLinkDB, gorm.Expr(conf.FileLinkDB+"+?", 1)).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CopyUserFile increase link err:")
		}
		// 查询文件大小
		file := &models.File{}
		if err := tx.Table(conf.FilePoolTB).Where(conf.FileUuidDB+"=?", FileUuid).
			Select(conf.FileSizeDB).Find(file).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CopyUserFile get user volume err:")
		}
		// 更改用户空间大小
		if err := tx.Table(conf.UserTB).Where(conf.UserUuidDB+"=?", src_file.UserUuid).
			Update(conf.UserNowVolumeDB, gorm.Expr(conf.UserNowVolumeDB+"+?", file.Size)).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CopyUserFile update user volume err:")
		}
		return nil
	}); err != nil {
		return -1, err
	}
	// 返回新文件id，用于文件夹复制
	id := copy_file.ID
	return int(id), nil
}

// 移动用户空间文件
func (d *DBClientImpl) UpdateUserFileParent(src_id, des_ParentId int) error {
	if err := d.DBConn.Table(conf.UserFileTB).Where(conf.UserFileIdDB+"=?", src_id).
		Update(conf.UserFileParentDB, des_ParentId).Error; err != nil {
		return errors.Wrap(err, "[DBClientImpl] UpdateUserFileParent update parent id err:")
	}
	return nil
}

// 文件名称修改
func (d *DBClientImpl) UpdateUserFileName(name, ext, uuid string) error {
	if err := d.DBConn.Table(conf.UserFileTB).Where(conf.UserFileUuidDB+"=?", uuid).
		Updates(models.UserFile{Name: name, Ext: ext}).Error; err != nil {
		return errors.Wrap(err, "[DBClientImpl] UpdateUserFileName update name err:")
	}
	return nil
}

// 删除单个用户文件，用于移动和删除
func (d *DBClientImpl) DeleteUserFileByUuid(UserFileUuid, FileUuid string) error {
	err := d.DBConn.Transaction(func(tx *gorm.DB) error {
		user_file := &models.UserFile{}
		// 先查询user_file，mysql不支持returning
		err := tx.Table(conf.UserFileTB).
			Where(conf.UserFileUuidDB+"=?", UserFileUuid).Find(user_file).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] DeleteUserFileByUuid get user file err:")
		}
		// 删除
		err = tx.Table(conf.UserFileTB).
			Where(conf.UserFileUuidDB+"=?", UserFileUuid).Delete(user_file).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] DeleteUserFileByUuid delete user file err:")
		}
		// 如果为文件夹则只删除user_file表中记录
		if FileUuid == "" {
			return nil
		}
		// 如果是文件则还需修改引用指针数量
		err = tx.Table(conf.FilePoolTB).Where(conf.FileUuidDB+"=?", FileUuid).
			Update(conf.FileLinkDB, gorm.Expr(conf.FileLinkDB+"+?", -1)).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] DeleteUserFileByUuid update link err:")
		}
		// 还要更新用户空间大小
		if err := tx.Table(conf.UserTB).Where(conf.UserUuidDB+"=?", user_file.UserUuid).
			Update(conf.UserNowVolumeDB, gorm.Expr(conf.UserNowVolumeDB+"-?", user_file.Size)).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] DeleteUserFileByUuid Update user err:")
		}
		return nil
	})
	return err
}

// 删除批量用户文件，用于批量移动和删除
// 弃用
func (d *DBClientImpl) DeleteUserFileBatch(uuids string) error {
	file := &models.UserFile{}
	err := d.DBConn.Table(conf.UserFileTB).Where(conf.FileUuidDB+" in (?)", uuids).Delete(file).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] DeleteUserFileBatch err:")
	}
	return nil
}

// 查看文件引用数
func (d *DBClientImpl) GetFileLink(uuid string) (link int, err error) {
	file := &models.File{}
	err = d.DBConn.Table(conf.FilePoolTB).Where(conf.FileUuidDB+"=?", uuid).Select(conf.UserFileIdDB).Find(file).Error
	if err != nil {
		return 0, errors.Wrap(err, "[DBClientImpl] GetFileLink err:")
	}
	return file.Link, nil
}

// 修改文件文件引用数
func (d *DBClientImpl) UpdateFileLink(uuid string, data int) error {
	err := d.DBConn.Table(conf.FilePoolTB).Where(conf.FileUuidDB+"=?", uuid).
		Update(conf.FileLinkDB, gorm.Expr(conf.FileLinkDB+"+?", data)).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] GetFileLink err:")
	}
	return nil
}

// 获取file_pool
func (d *DBClientImpl) GetFileByUuid(uuid string) (file *models.File, err error) {
	file = &models.File{}
	err = d.DBConn.Table(conf.FilePoolTB).Where(conf.FileUuidDB+"=?", uuid).First(file).Error
	if err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetFileByUuid err:")
	}
	return file, nil
}

// 通过fileKey查询file
func (d *DBClientImpl) GetFileByFileKey(fileKey string) (file *models.File, err error) {
	file = &models.File{}
	err = d.DBConn.Table(conf.FilePoolTB).Where(conf.FileFileKeyDB+"=?", fileKey).First(file).Error
	if err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetFileByFileKey err:")
	}
	return file, nil
}

// 通过UserFileUuid获取fileKey
func (d *DBClientImpl) GetFileKeyByUserFileUuid(uuid string) (fileKey string, err error) {
	file := &models.File{}
	// 拼接sql
	ft := conf.FilePoolTB
	fid := conf.FileUuidDB
	uft := conf.UserFileTB
	ufid := conf.UserFilePoolUuidDB
	// select "path" from "file_pool" inner join "user_file" on ("user_file".FileUuid = "file_pool".uuid) where "user_file".uuid = uuid
	err = d.DBConn.Table(conf.FilePoolTB).Joins(fmt.Sprintf("inner join %s on %s.%s = %s.%s", uft, ft, fid, uft, ufid)).
		Where(fmt.Sprintf("%s.%s=?", uft, conf.UserFileUuidDB), uuid).Select(conf.FileFileKeyDB).First(file).Error
	if err != nil {
		return "", errors.Wrap(err, "[DBClientImpl] GetFileKeyByUserFileUuid err:")
	}
	return file.FileKey, nil
}

func (d *DBClientImpl) CreateUserFile(user_file *models.UserFile) error {
	if err := d.DBConn.Table(conf.UserFileTB).Create(user_file).Error; err != nil {
		return errors.Wrap(err, "[DBClientImpl] CreateUserFile err:")
	}
	return nil
}

// 更新文件存储类型
func (d *DBClientImpl) UpdateFileStoreTypeByHash(hash string, t int) error {
	err := d.DBConn.Table(conf.FilePoolTB).Where(conf.FileHashDB+"=?", hash).
		Update(conf.FileStoreTypeDB, t).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] UpdateFileStoreType err:")
	}
	return nil
}

// 创建分享链接
func (d *DBClientImpl) SetShare(share *models.Share) error {
	err := d.DBConn.Transaction(func(tx *gorm.DB) error {
		// 先删除原来的
		err := tx.Table(conf.ShareTB).Where(conf.ShareUuidDB+"=?", share.Uuid).Delete(share).Error
		// 如果发生其他错误
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Wrap(err, "[DBClientImpl] SetShare delete err:")
		}
		// 没找到或成功删除则插入新的
		err = tx.Table(conf.ShareTB).Create(share).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] SetShare create err:")
		}
		return err
	})
	return err
}

// 获取share
func (d *DBClientImpl) GetShareByUuid(uuid string) (*models.Share, error) {
	share := &models.Share{}
	err := d.DBConn.Table(conf.ShareTB).Where(conf.ShareUuidDB+"=?", uuid).First(share).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, conf.DBNotFoundError
		}
		return nil, errors.Wrap(err, "[DBClientImpl] GetShareByUuid err:")
	}
	return share, nil
}

// 获取UserFileUuid
func (d *DBClientImpl) GetUserFileUuidByShareUuid(uuid string) (UserFileUuid string, err error) {
	share := &models.Share{}
	err = d.DBConn.Table(conf.ShareTB).Where(conf.ShareUuidDB).Select(conf.ShareUserFileUuidDB).First(share).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", conf.DBNotFoundError
		}
		return "", errors.Wrap(err, "[DBClientImpl] GetShareByUuid err:")
	}
	return share.UserFileUuid, nil
}

// 更新share
func (d *DBClientImpl) UpdateShareByUuid(uuid string, share *models.Share) error {
	// map无论是否为0值都会更新
	updateMap := map[string]interface{}{
		"code":       share.Code,
		"ExpireTime": share.ExpireTime,
	}
	err := d.DBConn.Model(&share).Updates(updateMap).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] UpdateShareByUuid err:")
	}
	return nil
}

// 获取share列表
func (d *DBClientImpl) GetShareListByUser(UserUuid string, cur, pageSize, mod int) (shares []*models.Share, err error) {
	// limit + offset实现分页查询
	switch mod {
	case conf.ShareAllMod:
		err = d.DBConn.Table(conf.ShareTB).
			Where(conf.ShareUserUuidDB+"=?", UserUuid).
			Limit(pageSize).Offset((cur - 1) * pageSize).Find(&shares).Error
	case conf.ShareExpireMod:
		err = d.DBConn.Table(conf.ShareTB).
			Where(conf.ShareUserUuidDB+"=?", UserUuid).
			Where(conf.ShareExpireDB+">=?", time.Now()).
			Limit(pageSize).Offset((cur - 1) * pageSize).Find(&shares).Error
	case conf.ShareOutMod:
		err = d.DBConn.Table(conf.ShareTB).
			Where(conf.ShareUserUuidDB+"=?", UserUuid).
			Where(conf.ShareExpireDB+"<?", time.Now()).
			Limit(pageSize).Offset((cur - 1) * pageSize).Find(&shares).Error
	}
	// 空说明错误
	if shares == nil || err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetShareListByUser err:")
	}
	return
}

// 删除share
func (d *DBClientImpl) DeleteShareByUuid(uuid string) error {
	share := &models.Share{}
	err := d.DBConn.Table(conf.ShareTB).Where(conf.ShareUuidDB+"=?", uuid).Delete(share).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] DeleteShareByUuid err:")
	}
	return nil
}

// 通过userfile删除share
func (d *DBClientImpl) DeleteShareByUserFileUuid(uuid string) error {
	share := &models.Share{}
	err := d.DBConn.Table(conf.ShareTB).Where(conf.ShareUserFileUuidDB+"=?", uuid).Delete(share).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] DeleteShareByUuid err:")
	}
	return nil
}

// 创建传输记录
func (d *DBClientImpl) CreateTrans(trans *models.Trans) error {
	err := d.DBConn.Table(conf.TransTB).Create(trans).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] CreateTrans err:")
	}
	return nil
}

// 更改trans记录类型
func (d *DBClientImpl) UpdateTransState(uuid string, state int) error {
	err := d.DBConn.Table(conf.TransTB).Where(conf.TransUuidDB+"=?", uuid).
		Update(conf.TransStatusDB, state).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] UpdateTransState Update state err:")
	}
	return nil
}

func (d *DBClientImpl) GetTransStatusByUuid(uuid string) (state int, err error) {
	trans := &models.Trans{}
	err = d.DBConn.Table(conf.TransTB).Where(conf.TransUuidDB+"=?", uuid).
		Find(trans).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return conf.TransNil, nil
	}
	if err != nil {
		return -2, errors.Wrap(err, "[DBClientImpl] GetTransStatusByUuid get state err:")
	}
	return trans.Status, nil
}

func (d *DBClientImpl) GetTransListByUser(UserUuid string, cur, pageSize, mod, status int) (trans []*models.Trans, err error) {
	// limit + offset实现分页查询
	err = d.DBConn.Table(conf.TransTB).
		Where(conf.TransUserUuidDB+"=?", UserUuid).
		Where(conf.TransIsDownDB+"=?", mod).
		Where(conf.TransStatusDB+"=?", status).
		Limit(pageSize).Offset((cur - 1) * pageSize).Find(&trans).Error
	if trans == nil || err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetTransListByUser err:")
	}
	return trans, nil
}

// 删除指定trans记录
func (d *DBClientImpl) DelTransByUuid(uuid string) error {
	trans := &models.Trans{}
	err := d.DBConn.Table(conf.TransTB).Where(conf.TransUuidDB+"=?", uuid).Delete(trans).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] DelTransByUuid err:")
	}
	return nil
}

// 根据状态和类别批量删除删除
func (d *DBClientImpl) DelTransByStatus(UserUuid string, mod, status int) error {
	trans := &models.Trans{}
	err := d.DBConn.Table(conf.TransTB).
		Where(conf.TransUserUuidDB+"=?", UserUuid).
		Where(conf.TransIsDownDB+"=?", mod).
		Where(conf.TransStatusDB+"=?", status).
		Delete(trans).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] DelTransByUuid err:")
	}
	return nil
}
