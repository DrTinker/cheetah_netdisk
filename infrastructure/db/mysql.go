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
func (d *DBClientImpl) CreateUploadRecord(file *models.File, userFile *models.UserFile, now int64) error {
	d.DBConn.Transaction(func(tx *gorm.DB) error {
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
		// 更改user表当前空间
		cur := now + int64(file.Size)
		if err := tx.Table(conf.User_TB).Where(conf.User_UUID_DB+"=?", userFile.User_Uuid).
			Update(conf.User_Now_Volume_DB, cur).Error; err != nil {
			return errors.Wrap(err, "[DBClientImpl] CreateUploadRecord Update user err:")
		}
		// 返回 nil 提交事务
		return nil
	})
	return nil
}
