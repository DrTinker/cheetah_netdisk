package helper

import (
	"NetDisk/conf"
	"fmt"
)

// 生成验证码rediskey
func GenVerifyCodeKey(prefix, email string) string {
	return fmt.Sprintf("%s_%s", prefix, email)
}
func GenUploadPartInfoKey(id string) string {
	return fmt.Sprintf("%s_%s", conf.UploadPartInfoKey, id)
}

func GenDownloadPartInfoKey(id string) string {
	return fmt.Sprintf("%s_%s", conf.DownloadPartInfoKey, id)
}
