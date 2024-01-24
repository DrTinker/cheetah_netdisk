package service

import (
	"NetDisk/client"
	"NetDisk/conf"

	"github.com/pkg/errors"
)

func GetPresignedURL(fileKey string) (url string, err error) {

	// 读取file_pool表获取当前文件所在存储位置：los or cos
	file, err := client.GetDBClient().GetFileByFileKey(fileKey)
	if err != nil {
		return "", errors.Wrap(err, "[GetPresignedURL] get file info error: ")
	}
	// 获取当前位置
	storeType := file.StoreType
	switch storeType {
	case conf.StoreTypeCOS:
		// cos的url需要增加domain
		url, err := client.GetCOSClient().GetPresignedUrl(fileKey, conf.DefaultSignExpire)
		if err != nil {
			return "", errors.Wrap(err, "[GetPresignedURL] get url error: ")
		}
		cfg, err := client.GetConfigClient().GetCOSConfig()
		if err != nil {
			return "", errors.Wrap(err, "[GetPresignedURL] get cos config error: ")
		}
		url = cfg.Domain + url
		return url, nil
	case conf.StoreTypeLOS:
		// minio返回的url直接能用
		url, err := client.GetLOSClient().GetPresignedUrl(fileKey, conf.DefaultSignExpire)
		if err != nil {
			return "", errors.Wrap(err, "[GetPresignedURL] get url error: ")
		}
		return url, nil
	}
	return "", nil
}
