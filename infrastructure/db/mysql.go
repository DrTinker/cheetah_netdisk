package db

import (
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
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
	sqlDB.SetMaxOpenConns(conf.Max_Conn)
	sqlDB.SetMaxIdleConns(conf.Max_Idle_Conn)
	sqlDB.SetConnMaxIdleTime(conf.Max_Idle_Time)
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
		err := tx.Table(conf.User_TB).Create(user).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] CreateUser Create user err:")
		}

		err = tx.Table(conf.User_File_TB).Create(folder).Error
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
	err := d.DBConn.Table(conf.User_TB).Where(conf.User_UUID_DB+"=?", id).First(user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(err, "[DBClientImpl] GetUserByID Select err:")
	}

	return user, nil
}

// 通过邮箱检索
func (d *DBClientImpl) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	// 查询用户信息
	err := d.DBConn.Table(conf.User_TB).Where(conf.User_Email_DB+"=?", email).First(user).Error
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
	err = d.DBConn.Table(conf.User_TB).Where(conf.User_UUID_DB+"=?", id).First(user).Error
	if err != nil {
		return 0, 0, errors.Wrap(err, "[DBClientImpl] GetUserVolume Select err:")
	}
	return user.Now_Volume, user.Total_Volume, nil
}

// 文件
// 检测文件存在性,false表示不存在
func (d *DBClientImpl) CheckFileExist(hash string) (bool, string, error) {
	file := &models.File{}
	err := d.DBConn.Table(conf.File_Pool_TB).Where(conf.File_Hash_DB+"=?", hash).Select(conf.File_UUID_DB).First(file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", nil
		}
		return false, "", errors.Wrap(err, "[DBClientImpl] CheckFileExist Select err:")
	}
	return true, file.Uuid, nil
}

func (d *DBClientImpl) CheckUserFileExist(user_uuid, file_uuid string) (bool, error) {
	user_file := &models.UserFile{}
	err := d.DBConn.Table(conf.User_File_TB).Where(&models.UserFile{File_Uuid: file_uuid, User_Uuid: user_uuid}).
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
		// user_file表增加记录
		if err := tx.Table(conf.User_File_TB).Create(userFile).Error; err != nil {
			// 返回任何错误都会回滚事务
			return errors.Wrap(err, "[DBClientImpl] CreateUploadRecord Create user file err:")
		}
		// file_pool表增加记录
		if err := tx.Table(conf.File_Pool_TB).Create(file).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CreateUploadRecord Create file err:")
		}
		// 检查文件大小
		user := &models.User{}
		err := tx.Table(conf.User_TB).Where(conf.User_UUID_DB+"=?", userFile.User_Uuid).First(user).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] GetUserVolume Select err:")
		}
		cur := user.Now_Volume + int64(file.Size)
		if user.Total_Volume < cur+int64(file.Size) {
			return conf.VolumeError
		}
		// 更新用户空间大小
		if err := tx.Table(conf.User_TB).Where(conf.User_UUID_DB+"=?", userFile.User_Uuid).
			Update(conf.User_Now_Volume_DB, cur).Error; err != nil {
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
		if err := tx.Table(conf.User_File_TB).Create(userFile).Error; err != nil {
			// 返回任何错误都会回滚事务
			return errors.Wrap(err, "[DBClientImpl] CreateQuickUploadRecord Create user file err:")
		}
		// 检查文件大小
		user := &models.User{}
		err := tx.Table(conf.User_TB).Where(conf.User_UUID_DB+"=?", userFile.User_Uuid).First(user).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] CreateQuickUploadRecord Get volume err:")
		}
		cur := user.Now_Volume + int64(size)
		if user.Total_Volume < cur+int64(size) {
			return conf.VolumeError
		}
		// 更新用户空间大小
		if err := tx.Table(conf.User_TB).Where(conf.User_UUID_DB+"=?", userFile.User_Uuid).
			Update(conf.User_Now_Volume_DB, cur).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CreateQuickUploadRecord Update user err:")
		}
		// 修改file_pool中文件引用数
		err = d.DBConn.Table(conf.File_Pool_TB).Where(conf.File_UUID_DB+"=?", userFile.File_Uuid).
			Update(conf.File_Link_DB, gorm.Expr(conf.File_Link_DB+"+?", 1)).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] CreateQuickUploadRecord Update link err:")
		}
		// 返回 nil 提交事务
		return nil
	})
	return err
}

// 删除上传记录，用于回滚
func (d *DBClientImpl) DeleteUploadRecord(file_uuid, user_file_uuid string) error {
	err := d.DBConn.Transaction(func(tx *gorm.DB) error {
		userFile := &models.UserFile{}
		// user_file表删除记录，软删除
		if err := tx.Table(conf.User_File_TB).Where(conf.User_File_UUID_DB+"=?", user_file_uuid).Delete(userFile).Error; err != nil {
			// 返回任何错误都会回滚事务
			return errors.Wrap(err, "[DBClientImpl] DeleteUploadRecord delete user file err:")
		}
		file := &models.File{}
		// file_pool表删除记录，真删除，与cos保持同步，当cos中文件被删除后用户空间无法恢复
		if err := tx.Table(conf.File_Pool_TB).Where(conf.File_UUID_DB+"=?", file_uuid).Unscoped().Delete(file).Error; err != nil {
			// 返回任何错误都会回滚事务
			return errors.Wrap(err, "[DBClientImpl] DeleteUploadRecord delete user file err:")
		}
		return nil
	})
	return err
}

// 获取父级文件
func (d *DBClientImpl) GetUserFileParent(uuid string) (file *models.UserFile, err error) {
	file = &models.UserFile{}
	err = d.DBConn.Table(conf.User_File_TB).Where(conf.User_File_UUID_DB+"=?", uuid).Find(file).Error
	if err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetUserFileParent err:")
	}
	return file, nil
}

// 获取文件列表
func (d *DBClientImpl) GetUserFileList(parent_id int) (files []*models.UserFile, err error) {
	// find无需初始化，且应传入数组指针
	err = d.DBConn.Table(conf.User_File_TB).Where(conf.User_File_Parent_DB+"=?", parent_id).Find(&files).Error
	if files == nil || err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetFileList err:")
	}
	return
}

// 通过文件uuid获取id
func (d *DBClientImpl) GetUserFileIDByUuid(uuids []string) (ids map[string]int, err error) {
	var files []models.UserFile
	ids = make(map[string]int, len(uuids))
	// 查出来file是乱序的
	err = d.DBConn.Table(conf.User_File_TB).Where(conf.File_UUID_DB+" in (?)", uuids).
		Select(conf.User_File_ID_DB, conf.User_File_UUID_DB).Find(&files).Error
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
	ft := conf.File_Pool_TB
	fid := conf.File_UUID_DB
	uft := conf.User_File_TB
	ufid := conf.User_File_Pool_UUID_DB
	// select * from "user_file" inner join "file_pool" on ("user_file".file_uuid = "file_pool".uuid) where "file_pool".path = path
	err = d.DBConn.Table(conf.User_File_TB).Joins(fmt.Sprintf("inner join %s on %s.%s = %s.%s", ft, ft, fid, uft, ufid)).
		Where(fmt.Sprintf("%s.%s=?", ft, conf.File_Path_DB), path).First(user_file).Error
	if err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetUserFileByPath err:")
	}
	return user_file, nil
}

// 获取单个用户文件信息，用于复制和移动
func (d *DBClientImpl) GetUserFileByUuid(uuid string) (file *models.UserFile, err error) {
	file = &models.UserFile{}
	err = d.DBConn.Table(conf.User_File_TB).Where(conf.File_UUID_DB+"=?", uuid).First(file).Error
	if err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetUserFileByUuid err:")
	}
	return file, nil
}

// 获取批量用户文件信息，用于批量复制和移动
func (d *DBClientImpl) GetUserFileBatch(uuids []string) (files []*models.UserFile, err error) {
	files = make([]*models.UserFile, len(uuids))
	err = d.DBConn.Table(conf.User_File_TB).Where(conf.File_UUID_DB+" in (?)", uuids).Find(&files).Error
	if err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetUserFileBatch err:")
	}
	return files, nil
}

func (d *DBClientImpl) GetUserByFileUuid(file_uuid string) (user_uuid string, err error) {
	user_file := &models.UserFile{}
	err = d.DBConn.Table(conf.User_File_TB).Where(conf.User_File_Pool_UUID_DB+"=?", file_uuid).
		Select(conf.User_File_User_ID_DB).Find(user_file).Error
	if err != nil {
		return "", errors.Wrap(err, "[DBClientImpl] GetUserByFileUuid err:")
	}
	return user_file.User_Uuid, nil
}

// 通过user_file_uuid查询对应的file_uuid
func (d *DBClientImpl) GetFileUuidByUserFileUuid(user_file_uuid string) (file_uuid string, err error) {
	user_file := &models.UserFile{}
	err = d.DBConn.Table(conf.User_File_TB).Where(conf.User_File_UUID_DB+"=?", user_file_uuid).
		Select(conf.User_File_Pool_UUID_DB).Find(user_file).Error
	if err != nil {
		return "", errors.Wrap(err, "[DBClientImpl] GetFileUuidByUserFileUuid err:")
	}
	return user_file.File_Uuid, nil
}

// 在用户文件空间复制
func (d *DBClientImpl) CopyUserFile(src_file *models.UserFile, des_parent_id int) (int, error) {
	// 生成新id, uuid和parentId
	copy_file := &models.UserFile{
		Uuid:      helper.GenUserFid(src_file.User_Uuid, src_file.Name+"_copy"),
		Parent_Id: des_parent_id,
		Name:      src_file.Name,
		Ext:       src_file.Ext,
		User_Uuid: src_file.User_Uuid,
		File_Uuid: src_file.File_Uuid,
	}
	// 复制文件
	if err := d.DBConn.Transaction(func(tx *gorm.DB) error {
		// 复制user_file记录
		if err := tx.Table(conf.User_File_TB).Create(copy_file).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CopyUserFile create copy err:")
		}
		// 增加file_pool中link数
		file_uuid := src_file.File_Uuid
		if err := tx.Table(conf.File_Pool_TB).Where(conf.File_UUID_DB+"=?", file_uuid).
			Update(conf.File_Link_DB, gorm.Expr(conf.File_Link_DB+"+?", 1)).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CopyUserFile increase link err:")
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
func (d *DBClientImpl) UpdateUserFileParent(src_id, des_parent_id int) error {
	if err := d.DBConn.Table(conf.User_File_TB).Where(conf.User_File_ID_DB+"=?", src_id).
		Update(conf.User_File_Parent_DB, des_parent_id).Error; err != nil {
		return errors.Wrap(err, "[DBClientImpl] UpdateUserFileParent update parent id err:")
	}
	return nil
}

// 文件名称修改
func (d *DBClientImpl) UpdateUserFileName(name, ext, uuid string) error {
	if err := d.DBConn.Table(conf.User_File_TB).Where(conf.User_File_UUID_DB+"=?", uuid).
		Updates(models.UserFile{Name: name, Ext: ext}).Error; err != nil {
		return errors.Wrap(err, "[DBClientImpl] UpdateUserFileName update name err:")
	}
	return nil
}

// 删除单个用户文件，用于移动和删除
func (d *DBClientImpl) DeleteUserFileByUuid(user_file_uuid, file_uuid string) error {
	err := d.DBConn.Transaction(func(tx *gorm.DB) error {
		file := &models.UserFile{}
		// 返回被删除的行
		err := d.DBConn.Table(conf.User_File_TB).Where(conf.File_UUID_DB+"=?", user_file_uuid).Delete(file).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] DeleteUserFileByUuid delete user file err:")
		}
		// 如果为文件夹则只删除user_file表中记录
		if file_uuid == "" {
			return nil
		}
		// 如果是文件则还需修改引用指针数量
		err = d.DBConn.Table(conf.File_Pool_TB).Where(conf.File_UUID_DB+"=?", file_uuid).
			Update(conf.File_Link_DB, gorm.Expr(conf.File_Link_DB+"+?", -1)).Error
		if err != nil {
			return errors.Wrap(err, "[DBClientImpl] DeleteUserFileByUuid update link err:")
		}
		return nil
	})
	return err
}

// 删除批量用户文件，用于批量移动和删除
// 弃用
func (d *DBClientImpl) DeleteUserFileBatch(uuids string) error {
	file := &models.UserFile{}
	err := d.DBConn.Table(conf.User_File_TB).Where(conf.File_UUID_DB+" in (?)", uuids).Delete(file).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] DeleteUserFileBatch err:")
	}
	return nil
}

// 查看文件引用数
func (d *DBClientImpl) GetFileLink(uuid string) (link int, err error) {
	file := &models.File{}
	err = d.DBConn.Table(conf.File_Pool_TB).Where(conf.File_UUID_DB+"=?", uuid).Select(conf.User_File_ID_DB).Find(file).Error
	if err != nil {
		return 0, errors.Wrap(err, "[DBClientImpl] GetFileLink err:")
	}
	return file.Link, nil
}

// 修改文件文件引用数
func (d *DBClientImpl) UpdateFileLink(uuid string, data int) error {
	err := d.DBConn.Table(conf.File_Pool_TB).Where(conf.File_UUID_DB+"=?", uuid).
		Update(conf.File_Link_DB, gorm.Expr(conf.File_Link_DB+"+?", data)).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] GetFileLink err:")
	}
	return nil
}

// 获取file_pool
func (d *DBClientImpl) GetFileByUuid(uuid string) (file *models.File, err error) {
	file = &models.File{}
	err = d.DBConn.Table(conf.File_Pool_TB).Where(conf.File_UUID_DB+"=?", uuid).First(file).Error
	if err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetFileByUuid err:")
	}
	return file, nil
}

func (d *DBClientImpl) CreateUserFile(user_file *models.UserFile) error {
	if err := d.DBConn.Table(conf.User_File_TB).Create(user_file).Error; err != nil {
		return errors.Wrap(err, "[DBClientImpl] CreateUserFile err:")
	}
	return nil
}

// 更新文件存储类型
func (d *DBClientImpl) UpdateFileStoreTypeByHash(hash string, t int) error {
	err := d.DBConn.Table(conf.File_Pool_TB).Where(conf.File_Hash_DB+"=?", hash).
		Update(conf.File_Store_Type_DB, t).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] UpdateFileStoreType err:")
	}
	return nil
}

// 创建分享链接
func (d *DBClientImpl) CreateShare(share *models.Share) error {
	err := d.DBConn.Table(conf.Share_TB).Create(share).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] CreateShare err:")
	}
	return nil
}

// 批量创建
func (d *DBClientImpl) CreateShareBatch(shares []*models.Share) error {
	err := d.DBConn.Table(conf.Share_TB).CreateInBatches(shares, 100).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] CreateShareBatch err:")
	}
	return nil
}

// 获取share
func (d *DBClientImpl) GetShareByUuid(uuid string) (*models.Share, error) {
	share := &models.Share{}
	err := d.DBConn.Table(conf.Share_TB).Where(conf.Share_UUID_DB).First(share).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, conf.DBDeleteError
		}
		return nil, errors.Wrap(err, "[DBClientImpl] GetShareByUuid err:")
	}
	return share, nil
}

// 获取user_file_uuid
func (d *DBClientImpl) GetUserFileUuidByShareUuid(uuid string) (user_file_uuid string, err error) {
	share := &models.Share{}
	err = d.DBConn.Table(conf.Share_TB).Where(conf.Share_UUID_DB).Select(conf.Share_User_File_UUID_DB).First(share).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", conf.DBDeleteError
		}
		return "", errors.Wrap(err, "[DBClientImpl] GetShareByUuid err:")
	}
	return share.User_File_Uuid, nil
}

// 删除share
func (d *DBClientImpl) DeleteShareByUuid(uuid string) error {
	share := &models.Share{}
	err := d.DBConn.Table(conf.Share_TB).Where(conf.Share_UUID_DB+"=?", uuid).Delete(share).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] DeleteShareByUuid err:")
	}
	return nil
}

// TODO 修改clickNum
func (d *DBClientImpl) UpdateClickNumByUuid(uuid string) error {
	err := d.DBConn.Table(conf.Share_TB).Where(conf.Share_UUID_DB+"=?", uuid).
		Update(conf.Share_Click_Num_DB, gorm.Expr(conf.Share_Click_Num_DB+"+?", 1)).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] UpdateClickNumByUuid Update click err:")
	}
	return nil
}
