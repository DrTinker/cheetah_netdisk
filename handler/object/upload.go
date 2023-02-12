package object

import (
	"NetDisk/conf"
	"NetDisk/handler/general"
	"NetDisk/helper"
	"NetDisk/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 单个上传文件 or 创建文件夹
// TODO 先实现串行上传多个文件，再优化为多线程上传
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
	fileKey := c.GetString(conf.File_Name_Form_Key)
	hash := c.GetString(conf.File_Hash_Key)

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
	user_file_uuid := helper.GenUserFid(user_uuid, fileKey)
	// 上传
	err = general.UploadObject(&models.UploadObjectParams{
		FileKey:        fileKey,
		User_Uuid:      user_uuid,
		Parent:         user_file_uuid_parent,
		Hash:           hash,
		Size:           int(file.Size),
		File_Uuid:      file_uuid,
		User_File_Uuid: user_file_uuid,
	}, fd)
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