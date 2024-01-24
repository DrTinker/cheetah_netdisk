package start

import (
	"NetDisk/client"
	"NetDisk/conf"
)

func InitJWT() {
	// 初始化jwt
	client.EncryptionClient.JWTInit(conf.JWTExpireValue, []byte(conf.JWTKeyValue))
}
