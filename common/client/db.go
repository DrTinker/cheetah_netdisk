package client

import (
	"NetDesk/common/models"
	"sync"
)

type DBClient interface {
	// user
	CreateUser(user *models.User, folder *models.UserFile) error

	GetUserByID(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserVolume(id string) (now, total int64, err error)

	// user_file
	CreateUserFile(user_file *models.UserFile) error
	CreateQuickUploadRecord(userFile *models.UserFile, size int) error
	CheckUserFileExist(user_uuid, file_uuid string) (bool, error)

	GetUserFileList(parent_id int) (files []*models.UserFile, err error)
	GetUserFileParent(uuid string) (file *models.UserFile, err error)
	GetUserFileIDByUuid(uuids []string) (ids map[string]int, err error)
	GetUserFileByPath(path string) (file *models.UserFile, err error)
	GetUserFileByUuid(uuid string) (file *models.UserFile, err error)
	GetUserFileBatch(uuids []string) (files []*models.UserFile, err error)

	GetFileUuidByUserFileUuid(user_file_uuid string) (file_uuid string, err error)
	GetUserByFileUuid(file_uuid string) (user_uuid string, err error)

	DeleteUserFileByUuid(user_file_uuid, file_uuid string) error
	DeleteUserFileBatch(uuids string) error

	UpdateUserFileParent(src_id, des_parent_id int) error
	UpdateUserFileName(name, ext, uuid string) error

	// file_pool
	CheckFileExist(hash string) (bool, string, error)
	GetFileLink(uuid string) (link int, err error)
	GetFileByUuid(uuid string) (file *models.File, err error)
	GetFileKeyByUserFileUuid(uuid string) (fileKey string, err error)

	UpdateFileLink(uuid string, data int) error
	UpdateFileStoreTypeByHash(hash string, t int) error

	// share
	CreateShare(share *models.Share) error
	CreateShareBatch(shares []*models.Share) error

	GetShareByUuid(uuid string) (*models.Share, error)
	GetUserFileUuidByShareUuid(uuid string) (user_file_uuid string, err error)

	DeleteShareByUuid(uuid string) error

	UpdateClickNumByUuid(uuid string) error

	// general
	CopyUserFile(src_file *models.UserFile, des_parent_id int) (int, error)
	CreateUploadRecord(file *models.File, userFile *models.UserFile) error
	DeleteUploadRecord(file_uuid, user_file_uuid string) error
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