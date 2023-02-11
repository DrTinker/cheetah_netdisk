package object

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

// TODO 完善分片上传
func InitUploadPartHandler(c *gin.Context) {
	// 获取文件唯一KEY
	fileKey := c.PostForm(conf.File_Name_Form_Key)
	if fileKey == "" {
		log.Error("UploadHandler empty file key")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取用户ID
	var user_uuid string
	if idstr, f := c.Get(conf.User_ID); f {
		user_uuid = helper.Strval(idstr)
	}
	if user_uuid == "" {
		log.Error("InitUploadPartHandler uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 生成文件uuid
	fileId := helper.GenFid(fileKey)
	userFileId := helper.GenUserFid(user_uuid, fileKey)
	// 调用COS接口
	uploadId, err := client.GetCOSClient().InitMultipartUpload(fileKey, nil)
	if err != nil || uploadId == "" {
		log.Error("InitUploadPartHandler get upload id error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_UPLOAD_PART_INIT_CODE,
			"msg":  conf.UPLOAD_PART_INIT_FAIL_MESSAGE,
		})
		return
	}
	log.Info("InitUploadPartHandler success: ", fileId)
	c.JSON(http.StatusBadRequest, gin.H{
		"code":         conf.ERROR_UPLOAD_PART_INIT_CODE,
		"msg":          conf.UPLOAD_PART_INIT_FAIL_MESSAGE,
		"file_id":      fileId,
		"user_file_id": userFileId,
	})

}
