package helper

import (
	"NetDesk/common/conf"
	"fmt"
)

// 生成验证码rediskey
func GenVerifyCodeKey(prefix, email string) string {
	return fmt.Sprintf("%s_%s", prefix, email)
}

func GenUploadPartInfoKey(id string) string {
	return fmt.Sprintf("%s_%s", conf.Upload_Part_Info_Key, id)
}
