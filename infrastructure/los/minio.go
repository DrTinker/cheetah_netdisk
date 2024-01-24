package los

import (
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"bytes"
	"context"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type LOSClientImpl struct {
	MinIOClient *minio.Client
}

func NewLOSClientImpl(cfg *models.LOSConfig) (*LOSClientImpl, error) {
	endpoint := cfg.Endpoint
	accessKeyID := cfg.AccessKeyID
	secretAccessKey := cfg.SecretAccessKey
	useSSL := cfg.UseSSL
	// 初始化minio client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	// 检测默认bucket是否存在，不存在则创建
	found, err := minioClient.BucketExists(context.Background(), conf.DefaultLOSBucket)
	if err != nil {
		return nil, err
	}
	if !found {
		err = minioClient.MakeBucket(context.Background(), conf.DefaultLOSBucket,
			minio.MakeBucketOptions{
				Region: "cn-south-1",
				// 对象锁开启后会默认开启版本控制，remove时只会删除最新版本的object
				// 导致大量旧版本占用存储空间
				// 有悖于minio在本项目中作为暂存系统的初衷
				ObjectLocking: false,
			})
		if err != nil {
			return nil, err
		}
		// 设置生命周期
		config := lifecycle.NewConfiguration()
		config.Rules = []lifecycle.Rule{
			{
				ID:     "expire-bucket",
				Status: "Enabled",
				// 默认私有云存储一天
				Expiration: lifecycle.Expiration{
					Days: conf.DefaultLOSExpire,
				},
			},
		}
		err = minioClient.SetBucketLifecycle(context.Background(), conf.DefaultLOSBucket, config)
		if err != nil {
			return nil, err
		}
	}

	return &LOSClientImpl{
		MinIOClient: minioClient,
	}, nil
}

// 上传至LOS
// fileKey LOS中路径
func (l *LOSClientImpl) PutObject(data []byte, fileKey string) error {
	reader := bytes.NewReader(data)
	_, err := l.MinIOClient.PutObject(context.Background(), conf.DefaultLOSBucket, fileKey, reader, int64(len(data)),
		minio.PutObjectOptions{})
	if err != nil {
		return errors.Wrap(err, "[LOSClientImpl] PutObject put object error: ")
	}
	// log.Info("[LOSClientImpl] PutObject success: %v", uploadInfo)
	return nil
}

// 将本地文件上传LOS
// clear 是否清除本地文件 true: 清除
func (l *LOSClientImpl) FPutObject(fileKey, filePath string, clear bool) error {
	uploadInfo, err := l.MinIOClient.FPutObject(context.Background(), conf.DefaultLOSBucket, fileKey, filePath, minio.PutObjectOptions{})
	if err != nil {
		return errors.Wrap(err, "[LOSClientImpl] FPutObject put object error: ")
	}
	if clear {
		err = helper.DelFile(filePath)
		if err != nil {
			return errors.Wrap(err, "[LOSClientImpl] FPutObject clear file error: ")
		}
	}
	log.Info("[LOSClientImpl] FPutObject success: ", uploadInfo)
	return nil
}

// 直接下载到本地文件中
// filePath 本地磁盘路径
// fileKey LOS中路径
func (l *LOSClientImpl) FGetObject(fileKey, filePath string) error {
	err := l.MinIOClient.FGetObject(context.Background(), conf.DefaultLOSBucket, fileKey, filePath, minio.GetObjectOptions{})
	return errors.Wrap(err, "[LOSClientImpl] FGetObject get object error: ")
}

// 删除文件 fileKey=test/aa.jpg
// 删除目录 fileKey=test/dir/
func (l *LOSClientImpl) RemoveObject(fileKey string) error {
	opts := minio.RemoveObjectOptions{}
	err := l.MinIOClient.RemoveObject(context.Background(), conf.DefaultLOSBucket, fileKey, opts)
	return errors.Wrap(err, "[LOSClientImpl] RemoveObject remove object error: ")
}

func (l *LOSClientImpl) RemoveDir(dir string) error {
	objs := l.MinIOClient.ListObjects(context.TODO(), conf.DefaultLOSBucket, minio.ListObjectsOptions{
		Prefix:    dir,
		Recursive: true,
	})
	errChan := l.MinIOClient.RemoveObjects(context.TODO(), conf.DefaultLOSBucket, objs, minio.RemoveObjectsOptions{})
	// 有错误只提示，等后续手动去删
	for e := range errChan {
		if e.Err != nil {
			return errors.Wrap(e.Err, "[LOSClientImpl] RemoveDir err:")
		}
	}
	return nil
}

// 合并分片文件
// src 一个目录 test/tmp/
// des 一个文件 test/merge.jpg
// bool 清除分片
// func (l *LOSClientImpl) MergeObjects(src, des string, clear bool) error {
// 	objs := l.MinIOClient.ListObjects(context.TODO(), conf.DefaultLOSBucket, minio.ListObjectsOptions{
// 		Prefix:    src,
// 		Recursive: true,
// 	})
// 	// 存本地
// 	srcTmp := fmt.Sprintf("%s/%s", conf.DefaultLocalPrefix, src)
// 	desTmp := fmt.Sprintf("%s/%s", conf.DefaultLocalPrefix, des)
// 	i := 0
// 	for obj := range objs {
// 		l.MinIOClient.FGetObject(context.TODO(), conf.DefaultLOSBucket, )
// 		helper.WriteFile(srcTmp, strconv.Itoa(i), obj.)
// 	}
// 	_, err := helper.MergeFile(srcTmp, desTmp)
// 	desFlag, _ := helper.PathExists(des)
// 	// 如果分片文件夹不存在 且 目标文件也不存在
// 	if err != nil && !desFlag {
// 		return nil, "", errors.Wrap(err, "[CompleteUploadPart] merge file error: ")
// 	}
// 	// 删除分片文件夹
// 	err = helper.RemoveDir(src[:len(src)-1])
// 	if err != nil {
// 		logrus.Warn("[CompleteUploadPart] remove src err: ", err)
// 	}
// 	// 检查文件hash合法性
// 	hash := helper.CountMD5(des, nil, 0)
// 	if hash != param.Hash {
// 		// 直接改为失败
// 		client.GetDBClient().UpdateTransState(uploadID, conf.TransFail)
// 		// 删除合并后的文件
// 		helper.DelFile(des)
// 		return nil, "", conf.InvaildFileHashError
// 	}
// 	srcOpts := make([]minio.CopySrcOptions, 0)
// 	for obj := range objs {
// 		opt := minio.CopySrcOptions{
// 			Bucket: conf.DefaultLOSBucket,
// 			Object: obj.Key,
// 		}
// 		srcOpts = append(srcOpts, opt)
// 	}
// 	dstOpts := minio.CopyDestOptions{
// 		Bucket: conf.DefaultLOSBucket,
// 		Object: des,
// 	}
// 	// 合并
// 	uploadInfo, err := l.MinIOClient.ComposeObject(context.Background(), dstOpts, srcOpts...)
// 	if err != nil {
// 		return errors.Wrap(err, "[LOSClientImpl] MergeObjects compose object error: ")
// 	}
// 	// 走到这说明合并成功
// 	if clear {
// 		objs := l.MinIOClient.ListObjects(context.TODO(), conf.DefaultLOSBucket, minio.ListObjectsOptions{
// 			Prefix:    src,
// 			Recursive: true,
// 		})
// 		errChan := l.MinIOClient.RemoveObjects(context.TODO(), conf.DefaultLOSBucket, objs, minio.RemoveObjectsOptions{})
// 		// 有错误只提示，等后续手动去删
// 		for e := range errChan {
// 			if e.Err != nil {
// 				log.Warn("[LOSClientImpl] Clear upload parts err:", e.ObjectName)
// 			}
// 		}
// 	}
// 	log.Info("[LOSClientImpl] Composed object successfully:", uploadInfo.Key)
// 	return nil
// }

func (l *LOSClientImpl) MergeObjects(src, des, contentType string, clear bool) error {
	objs := l.MinIOClient.ListObjects(context.TODO(), conf.DefaultLOSBucket, minio.ListObjectsOptions{
		Prefix:    src,
		Recursive: true,
	})
	srcOpts := make([]minio.CopySrcOptions, 0)
	for obj := range objs {
		opt := minio.CopySrcOptions{
			Bucket: conf.DefaultLOSBucket,
			Object: obj.Key,
		}
		srcOpts = append(srcOpts, opt)
	}
	dstOpts := minio.CopyDestOptions{
		Bucket:          conf.DefaultLOSBucket,
		Object:          des,
		UserMetadata:    map[string]string{"Content-Type": contentType},
		ReplaceMetadata: true,
	}
	// 合并
	uploadInfo, err := l.MinIOClient.ComposeObject(context.Background(), dstOpts, srcOpts...)
	if err != nil {
		return errors.Wrap(err, "[LOSClientImpl] MergeObjects compose object error: ")
	}
	// 走到这说明合并成功
	if clear {
		objs := l.MinIOClient.ListObjects(context.TODO(), conf.DefaultLOSBucket, minio.ListObjectsOptions{
			Prefix:    src,
			Recursive: true,
		})
		errChan := l.MinIOClient.RemoveObjects(context.TODO(), conf.DefaultLOSBucket, objs, minio.RemoveObjectsOptions{})
		// 有错误只提示，等后续手动去删
		for e := range errChan {
			if e.Err != nil {
				log.Warn("[LOSClientImpl] Clear upload parts err:", e.ObjectName)
			}
		}
	}
	log.Info("[LOSClientImpl] Composed object successfully:", uploadInfo.Key)
	return nil
}

// 生成访问签名
func (l *LOSClientImpl) GetPresignedUrl(fileKey string, expire time.Duration) (sign string, err error) {
	reqParams := make(url.Values)

	// Generates a presigned url which expires in a day.
	presignedURL, err := l.MinIOClient.PresignedGetObject(context.Background(), conf.DefaultLOSBucket, fileKey, expire, reqParams)
	if err != nil {
		return "", errors.Wrap(err, "[LOSClientImpl] GetPresignedUrl error: ")
	}

	return presignedURL.String(), nil
}
