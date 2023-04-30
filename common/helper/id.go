package helper

import (
	"NetDesk/common/conf"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

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
	return fmt.Sprintf("%s/%s.%s", conf.Default_System_Prefix, hash, ext)
}

// 生成uploadID
func GenUploadID(user, hash string) string {
	return fmt.Sprintf("UP_%s_%s_%d", user, hash, time.Now().UnixNano())
}

// 生成share uuid
func GenSid(user, code string) string {
	u := uuid.NewV4()
	str := u.String() + user + code
	id := md5.Sum([]byte(str))

	return fmt.Sprintf("%x", id)
}

// 生成服务全局id
func GenServiceID(service string, port *int) string {
	var h [16]byte
	rand.Read(h[:])
	// 生成一个全局ID
	id := fmt.Sprintf("%s-%s-%d", service, hex.EncodeToString(h[:]), *port)
	return id
}

// 生成验证码rediskey
func GenVerifyCodeKey(prefix, email string) string {
	return fmt.Sprintf("%s_%s", prefix, email)
}
