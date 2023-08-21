package helper

import (
	"NetDesk/conf"
	"fmt"
)

// 生成验证码rediskey
func GenVerifyCodeKey(prefix, email string) string {
	return fmt.Sprintf("%s_%s", prefix, email)
}
func GenUploadPartInfoKey(id string) string {
	return fmt.Sprintf("%s_%s", conf.Upload_Part_Info_Key, id)
}

func GenDownloadPartInfoKey(id string) string {
	return fmt.Sprintf("%s_%s", conf.Download_Part_Info_Key, id)
}
