package cos

import (
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type COSClientImpl struct {
	COSClient *cos.Client
}

func NewCOSClientImpl(cfg *models.COSConfig) (*COSClientImpl, error) {
	u, err := url.Parse(cfg.Domain)
	if err != nil {
		return nil, err
	}
	// 用于 Get Service 查询，默认全地域 service.cos.myqcloud.com
	su, err := url.Parse(cfg.Region)
	if err != nil {
		return nil, err
	}
	b := &cos.BaseURL{BucketURL: u, ServiceURL: su}
	// 永久密钥
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretId,
			SecretKey: cfg.SecretKey,
		},
	})

	return &COSClientImpl{COSClient: client}, nil
}

// 上传本地磁盘文件， key为COS中路径，path为本地路径(绝对路径)
func (c *COSClientImpl) UpLoadLocalFile(key, path string) error {
	_, _, err := c.COSClient.Object.Upload(context.Background(), key, path, &cos.MultiUploadOptions{
		ThreadPoolSize: conf.DefaultThreadPoolSize,
	})
	if err != nil {
		return errors.Wrap(err, "[COSClientImpl] UpLoadLocalFile upload file error: ")
	}
	return nil
}

func (c *COSClientImpl) UploadStream(key string, stream io.Reader) error {
	_, err := c.COSClient.Object.Put(context.Background(), key, stream, nil)
	if err != nil {
		return errors.Wrap(err, "[COSClientImpl] UploadStream upload file error: ")
	}
	return nil
}

func (c *COSClientImpl) InitMultipartUpload(key string, opts *models.MultiFileUploadOptions) (string, error) {
	info, _, err := c.COSClient.Object.InitiateMultipartUpload(context.Background(), key, nil)
	if err != nil {
		return "", errors.Wrap(err, "[COSClientImpl] UpLoadFileStream init error: ")
	}
	log.Info("[COSClientImpl] UpLoadFileStream init: ", info)
	return info.UploadID, nil
}

func (c *COSClientImpl) CompleteMultipartUpload(key, uploadID string, tags []models.Part) error {
	// 调用分片传输完成
	opt := &cos.CompleteMultipartUploadOptions{}
	opt.Parts = make([]cos.Object, len(tags))
	for i, v := range tags {
		opt.Parts[i].ETag = v.ETag
		opt.Parts[i].PartNumber = v.PartNum
	}
	_, _, err := c.COSClient.Object.CompleteMultipartUpload(
		context.Background(), key, uploadID, opt,
	)
	if err != nil {
		return errors.Wrap(err, "[COSClientImpl] CompleteMultipartUpload complete upload error: ")
	}

	return nil
}

func (c *COSClientImpl) UploadPart(pos int, data []byte, key, uploadID string) (*models.Part, error) {
	r := bytes.NewReader(data)
	// 分片上传
	resp, err := c.COSClient.Object.UploadPart(context.Background(), key, uploadID, pos, r, nil)
	if err != nil {
		// 出错终止
		log.Error("[COSClientImpl] UpLoadFileStream upload part error: ", err)
		_, err := c.COSClient.Object.AbortMultipartUpload(context.Background(), key, uploadID)
		if err != nil {
			return nil, errors.Wrap(err, "[COSClientImpl] UpLoadFileStream abort upload error: ")
		}
	}
	PartETag := resp.Header.Get("ETag")
	tag := &models.Part{
		PartNum: pos,
		ETag:    PartETag,
	}
	return tag, nil
}

func (c *COSClientImpl) Copy(src, des string) error {
	// 创建原地址
	sourceURL := fmt.Sprintf(c.COSClient.BaseURL.BucketURL.String()+"/%s", src)
	_, _, err := c.COSClient.Object.Copy(context.Background(), des, sourceURL, nil)
	if err != nil {
		return errors.Wrap(err, "[COSClientImpl] Copy file error: ")
	}
	return nil
}

// 删除对象为文件夹时不影响其下子文件，全部子文件的KEY不变，可按原方式访问
func (c *COSClientImpl) Delete(key string) error {
	_, err := c.COSClient.Object.Delete(context.Background(), key)
	if err != nil {
		return errors.Wrap(err, "[COSClientImpl] Delete file error: ")
	}
	return nil
}

// 生成预签名URL，用于文件上传下载
func (c *COSClientImpl) GetPresignedUrl(fileKey string, expire time.Duration) (url string, err error) {
	secretID := c.COSClient.GetCredential().SecretID
	secretKey := c.COSClient.GetCredential().SecretKey
	presignedURL, err := c.COSClient.Object.GetPresignedURL(context.Background(), http.MethodGet, fileKey, secretID, secretKey, expire, nil)
	if err != nil {
		return "", errors.Wrap(err, "[COSClientImpl] GetPresignedUrl error: ")
	}
	return presignedURL.RequestURI(), err
}

// 下载COS文件至服务端磁盘
func (c *COSClientImpl) DownloadLocal(fileKey, path string) error {
	opt := &cos.MultiDownloadOptions{
		ThreadPoolSize: conf.DefaultThreadPoolSize,
	}
	_, err := c.COSClient.Object.Download(context.Background(), fileKey, path, opt)
	if err != nil {
		return errors.Wrap(err, "[COSClientImpl] DownloadLocal error: ")
	}
	return nil
}

// 上传文件流，超过16M时将进行分片并通过多线程上传，每片大小16M，当有分片上传失败时终止上传
// 弃用
func (c *COSClientImpl) UpLoadStreamPart(key string, stream io.Reader, opts *models.MultiFileUploadOptions) error {
	data, err := ioutil.ReadAll(stream)
	if err != nil {
		return errors.Wrap(err, "[COSClientImpl] UpLoadFileStreamPart parse data error: ")
	}
	// 为文件分片
	slices := helper.ArrayInGroupsOf(data, conf.FilePartSizeMax)
	batch := len(slices)
	log.Info("[COSClientImpl] UpLoadFileStream batch: ", batch, "total size: ", len(data))
	// 初始化分块上传
	info, _, err := c.COSClient.Object.InitiateMultipartUpload(context.Background(), key, nil)
	if err != nil {
		return errors.Wrap(err, "[COSClientImpl] UpLoadFileStream init error: ")
	}
	log.Info("[COSClientImpl] UpLoadFileStream init: ", info)
	// 完成上传结构体
	opt := &cos.CompleteMultipartUploadOptions{}
	opt.Parts = make([]cos.Object, batch)
	var wg sync.WaitGroup
	// 分片上传
	for i, v := range slices {
		wg.Add(1)
		go uploadPart(&wg, c.COSClient, i+1, v, key, info.UploadID, opt)
	}
	// 等待全部子线程运行完毕
	wg.Wait()
	// 调用分片传输完成
	_, _, err = c.COSClient.Object.CompleteMultipartUpload(
		context.Background(), key, info.UploadID, opt,
	)
	if err != nil {
		log.Error("[COSClientImpl] UpLoadFileStream complete upload error: ", err)
	}
	return nil
}

func uploadPart(wg *sync.WaitGroup, c *cos.Client, pos int, data []byte, key, uploadID string, opt *cos.CompleteMultipartUploadOptions) {
	defer wg.Done()
	r := bytes.NewReader(data)
	// 分片上传
	log.Info("[COSClientImpl] thread run part: ", pos)
	resp, err := c.Object.UploadPart(context.Background(), key, uploadID, pos, r, nil)
	if err != nil {
		// 出错终止
		log.Error("[COSClientImpl] UpLoadFileStream upload part error: ", err)
		_, err := c.Object.AbortMultipartUpload(context.Background(), key, uploadID)
		if err != nil {
			log.Error("[COSClientImpl] UpLoadFileStream abort upload error: ", err)
		}
	}
	PartETag := resp.Header.Get("ETag")
	log.Info("Etag: ", PartETag)
	opt.Parts[pos-1] = cos.Object{
		PartNumber: pos,
		ETag:       PartETag,
	}
}
