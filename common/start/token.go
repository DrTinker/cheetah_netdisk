package start

import (
	"NetDesk/common/client"
	"NetDesk/common/conf"
)

func InitJWT() {
	// 初始化jwt
	client.EncryptionClient.JWTInit(conf.JWTExpireValue, []byte(conf.JWTKeyValue))
}
