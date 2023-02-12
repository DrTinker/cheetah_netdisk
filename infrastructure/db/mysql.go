package db

import (
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"NetDisk/conf"
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
// 创建用户
func (d *DBClientImpl) CreateUser(user *models.User) error {
	err := d.DBConn.Table(conf.User_TB).Create(user).Error
	if err != nil {
		return errors.Wrap(err, "[DBClientImpl] CreateUser Create err:")
	}
	return nil
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
func (d *DBClientImpl) CheckFileExist(hash string) (bool, error) {
	file := &models.File{}
	err := d.DBConn.Table(conf.File_Pool_TB).Where(conf.File_Hash_DB+"=?", hash).First(file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, errors.Wrap(err, "[DBClientImpl] GetUserByEmail Select err:")
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

// 删除上传记录，用于回滚
func (d *DBClientImpl) DeleteUploadRecord(file_uuid, user_file_uuid string) error {
	err := d.DBConn.Transaction(func(tx *gorm.DB) error {
		userFile := &models.UserFile{}
		// user_file表增加记录
		if err := tx.Table(conf.User_File_TB).Where(conf.User_File_UUID_DB+"=?", user_file_uuid).Delete(userFile).Error; err != nil {
			// 返回任何错误都会回滚事务
			return errors.Wrap(err, "[DBClientImpl] DeleteUploadRecord delete user file err:")
		}
		file := &models.File{}
		// user_file表增加记录
		if err := tx.Table(conf.File_Pool_TB).Where(conf.File_UUID_DB+"=?", file_uuid).Delete(file).Error; err != nil {
			// 返回任何错误都会回滚事务
			return errors.Wrap(err, "[DBClientImpl] DeleteUploadRecord delete user file err:")
		}
		return nil
	})
	return err
}

// 获取文件列表
func (d *DBClientImpl) GetFileList(parentId int) (files []*models.UserFile, err error) {
	// find无需初始化，且应传入数组指针
	err = d.DBConn.Table(conf.User_File_TB).Where(conf.User_File_Parent_DB+"=?", parentId).Find(&files).Error
	if files == nil || err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetFileList err:")
	}
	return
}

// 通过文件uuid获取id
func (d *DBClientImpl) GetFileIDByUuid(uuids []string) (ids []int, err error) {
	var files []models.UserFile
	ids = make([]int, len(uuids))
	err = d.DBConn.Table(conf.User_File_TB).Where(conf.File_UUID_DB+" in (?)", uuids).Find(&files).Error
	if files == nil || err != nil {
		return nil, errors.Wrap(err, "[DBClientImpl] GetFileIDByUuid err:")
	}
	for i, file := range files {
		ids[i] = int(file.ID)
	}
	return ids, nil
}
