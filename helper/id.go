package helper

import (
	"NetDisk/conf"
	"crypto/md5"
	"fmt"

	uuid "github.com/satori/go.uuid"
)

// 生成用户id
func GenUid(name string, email string) string {
	u := uuid.NewV4()
	str := u.String() + name + email
	id := md5.Sum([]byte(str))

	return fmt.Sprintf("%x", id)
}

// 生成文件id
func GenFid(key string) string {
	u := uuid.NewV4()
	str := u.String() + key
	id := md5.Sum([]byte(str))

	return fmt.Sprintf("%x", id)
}

// 生成用户空间文件id
func GenUserFid(user, name string) string {
	u := uuid.NewV4()
	str := u.String() + user + name
	id := md5.Sum([]byte(str))

	return fmt.Sprintf("%x", id)
}

// 生成fileKey
func GenFileKey(hash, ext string) string {
	return fmt.Sprintf("%s/%s.%s", conf.FilePrefix, hash, ext)
}

func GenThumbnailKey(name string) string {
	if name == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", conf.ThumbnailPrefix, name)
}

// 生成uploadID
func GenUploadID(user, hash string) string {
	u := uuid.NewV4()
	str := u.String() + user + hash
	id := md5.Sum([]byte(str))

	return fmt.Sprintf("%x", id)
}

func GenDownloadID(user, UserFileUuid string) string {
	u := uuid.NewV4()
	str := u.String() + user + UserFileUuid
	id := md5.Sum([]byte(str))

	return fmt.Sprintf("%x", id)
}

// 生成share uuid
func GenSid(user, code string) string {
	u := uuid.NewV4()
	str := u.String() + user + code
	id := md5.Sum([]byte(str))

	return fmt.Sprintf("%x", id)
}
