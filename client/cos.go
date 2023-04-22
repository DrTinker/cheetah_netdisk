package client

import (
	"NetDesk/models"
	"io"
	"sync"
	"time"
)

type COSClient interface {
	// 上传
	UpLoadLocalFile(key, path string) error
	UpLoadStreamPart(key string, stream io.Reader, opts *models.MultiFileUploadOptions) error // 弃用
	UploadStream(key string, stream io.Reader) error
	// 复制
	Copy(src, des string) error
	// 分片上传
	InitMultipartUpload(key string, opts *models.MultiFileUploadOptions) (string, error)
	CompleteMultipartUpload(key, uploadID string, tags []models.Part) error
	UploadPart(pos int, data []byte, key, uploadID string) (*models.Part, error)
	// 删除
	Delete(key string) error
	// URL
	GetPresignedUrl(fileKey string, expire time.Duration) (url string, err error)
}

var (
	cos     COSClient
	COSOnce sync.Once
)

func GetCOSClient() COSClient {
	return cos
}

func InitCOSClient(client COSClient) {
	COSOnce.Do(
		func() {
			cos = client
		},
	)
}
