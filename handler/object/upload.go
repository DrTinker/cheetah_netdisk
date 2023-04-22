package object

import (
	"NetDesk/client"
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
	"NetDesk/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 单个上传文件
// TODO 文件
func UploadHandler(c *gin.Context) {
	// gin获取文件和文件key
	file, err := c.FormFile(conf.File_Form_Key)
	if err != nil {
		log.Error("UploadHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 检查存在性的中间件已经读取过了，因此从ctx中获取
	hash := c.GetString(conf.File_Hash_Key)
	// 文件夹名称
	fileName := c.PostForm(conf.File_Name_Key)
	name, ext, err := helper.SplitFileFullName(fileName)
	if err != nil {
		log.Error("UploadHandler invaild file name")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	fileKey := helper.GenFileKey(hash, ext)

	// 前端传入uuid后端查询id
	user_file_uuid_parent := c.PostForm(conf.Folder_Uuid_Key)
	if fileKey == "" || user_file_uuid_parent == "" {
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
		log.Error("UploadHandler uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 读取文件
	fd, err := file.Open()
	if err != nil {
		log.Error("UploadHandler file open err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 生成ID
	file_uuid := helper.GenFid(fileKey)
	// 查看是否秒传
	flag := c.GetBool(conf.File_Quick_Upload_Key)
	if flag {
		// 秒传file_pool中uuid不变
		file_uuid = c.GetString(conf.File_Uuid_Key)
	}
	user_file_uuid := helper.GenUserFid(user_uuid, fileKey)
	// 上传
	err = service.UploadObjectServer(&models.UploadObjectParams{
		FileKey:        fileKey,
		User_Uuid:      user_uuid,
		Parent:         user_file_uuid_parent,
		Hash:           hash,
		Size:           int(file.Size),
		Name:           name,
		Ext:            ext,
		File_Uuid:      file_uuid,
		User_File_Uuid: user_file_uuid,
	}, fd, flag)
	if err != nil {
		log.Error("UploadHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_UPLOAD_CODE,
			"msg":  fmt.Sprintf(conf.UPLOAD_FAIL_MESSAGE, file.Filename),
		})
		return
	}

	log.Info("UploadHandler success: ", file_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.SUCCESS_RESP_MESSAGE,
		"msg":     fmt.Sprintf(conf.UPLOAD_SUCCESS_MESSAGE),
		"file_id": user_file_uuid,
	})
}

// TODO 完善分片上传
func InitUploadPartHandler(c *gin.Context) {
	// 获取文件唯一KEY
	fileKey := c.PostForm(conf.File_Name_Key)
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
