package start

import (
	"NetDesk/client"
	"NetDesk/conf"
)

func InitJWT() {
	// 初始化jwt
	client.EncryptionClient.JWTInit(conf.JWTExpireValue, []byte(conf.JWTKeyValue))
}
