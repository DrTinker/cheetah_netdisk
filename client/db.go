package client

import (
	"NetDisk/models"
	"sync"
)

type DBClient interface {
	// user
	// 创建用户
	CreateUser(user *models.User, folder *models.UserFile) error
	// 通过uuid获取用户信息
	GetUserByID(id string) (*models.User, error)
	// 通过email获取用户
	GetUserByEmail(email string) (*models.User, error)
	// 通过uuid获取用户空间大小
	GetUserVolume(id string) (now, total int64, err error)
	// 更改用户信息
	UpdateUserName(uuid, name string) error

	// user_file
	// 创建用户文件
	CreateUserFile(user_file *models.UserFile) error
	// 创建秒传记录
	CreateQuickUploadRecord(userFile *models.UserFile, size int) error
	// 查询用户文件是否存在
	CheckUserFileExist(UserUuid, FileUuid string) (bool, error)

	GetUserFileList(ParentId int) (files []*models.UserFile, err error)
	GetUserFileListPage(ParentId int, cur, pageSize int, ext string) (files []*models.UserFile, err error)
	GetUserFileParent(uuid string) (file *models.UserFile, err error)
	GetUserFileIDByUuid(uuids []string) (ids map[string]int, err error)
	GetUserFileByPath(path string) (file *models.UserFile, err error)
	GetUserFileByUuid(uuid string) (file *models.UserFile, err error)
	GetUserFileBatch(uuids []string) (files []*models.UserFile, err error)

	GetFileUuidByUserFileUuid(UserFileUuid string) (FileUuid string, err error)
	GetUserByFileUuid(FileUuid string) (UserUuid string, err error)

	DeleteUserFileByUuid(UserFileUuid, FileUuid string) error
	DeleteUserFileBatch(uuids string) error

	UpdateUserFileParent(src_id, des_ParentId int) error
	UpdateUserFileName(name, ext, uuid string) error

	// file_pool
	CheckFileExist(hash string) (bool, *models.File, error)
	GetFileLink(uuid string) (link int, err error)
	GetFileByUuid(uuid string) (file *models.File, err error)
	GetFileByFileKey(fileKey string) (file *models.File, err error)
	GetFileKeyByUserFileUuid(uuid string) (fileKey string, err error)

	UpdateFileLink(uuid string, data int) error
	UpdateFileStoreTypeByHash(hash string, t int) error

	// share
	SetShare(share *models.Share) error

	GetShareListByUser(UserUuid string, cur, pageSize, mod int) ([]*models.Share, error)
	GetShareByUuid(uuid string) (*models.Share, error)
	GetUserFileUuidByShareUuid(uuid string) (UserFileUuid string, err error)

	UpdateShareByUuid(uuid string, share *models.Share) error

	DeleteShareByUuid(uuid string) error
	DeleteShareByUserFileUuid(uuid string) error

	// trans
	CreateTrans(trans *models.Trans) error
	UpdateTransState(uuid string, state int) error
	GetTransStatusByUuid(uuid string) (state int, err error)
	GetTransListByUser(UserUuid string, cur, pageSize, mod, status int) ([]*models.Trans, error)

	DelTransByUuid(uuid string) error
	DelTransByStatus(UserUuid string, mod, status int) error

	// general
	CopyUserFile(src_file *models.UserFile, des_ParentId int) (int, error)
	CreateUploadRecord(file *models.File, userFile *models.UserFile) error
	DeleteUploadRecord(FileUuid, UserFileUuid string) error
}

var (
	db     DBClient
	DBOnce sync.Once
)

func GetDBClient() DBClient {
	return db
}

func InitDBClient(client DBClient) {
	DBOnce.Do(
		func() {
			db = client
		},
	)
}
