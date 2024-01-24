package client

import (
	"sync"
	"time"
)

// 私有云 local object storage
type LOSClient interface {
	PutObject(data []byte, fileKey string) error
	FPutObject(fileKey, filePath string, clear bool) error
	FGetObject(fileKey, filePath string) error
	RemoveObject(fileKey string) error
	MergeObjects(src, des, contentType string, clear bool) error
	RemoveDir(dir string) error
	GetPresignedUrl(fileKey string, expire time.Duration) (sign string, err error)
}

var (
	los     LOSClient
	LOSOnce sync.Once
)

func GetLOSClient() LOSClient {
	return los
}

func InitLOSClientt(client LOSClient) {
	LOSOnce.Do(
		func() {
			los = client
		},
	)
}
