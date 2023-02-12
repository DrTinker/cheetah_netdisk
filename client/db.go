package client

import (
	"NetDisk/models"
	"sync"
)

type DBClient interface {
	// user
	CreateUser(user *models.User) error
	GetUserByID(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserVolume(id string) (now, total int64, err error)

	// file
	CheckFileExist(hash string) (bool, error)
	CreateUploadRecord(file *models.File, userFile *models.UserFile) error
	GetFileList(parentId int) (files []*models.UserFile, err error)
	GetFileIDByUuid(uuids []string) (ids []int, err error)
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
