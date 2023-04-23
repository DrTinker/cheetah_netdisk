package helper

import (
	"NetDesk/conf"
	"fmt"
	"time"
)

func GenEmailVerifyKey(user string) string {
	return fmt.Sprintf("%s_%s_%d", conf.Code_Cache_Key, user, time.Now().UnixNano())
}

func GenUploadPartInfoKey(id string) string {
	return fmt.Sprintf("%s_%s", conf.Upload_Part_Info_Key, id)
}
