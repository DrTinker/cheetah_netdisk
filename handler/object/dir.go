package object

import (
	"NetDisk/conf"
	"NetDisk/handler/general"
	"NetDisk/helper"
	"NetDisk/models"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func MakeDirHandler(c *gin.Context) {
	// 获取文件夹路径
	fileKey := c.GetString(conf.File_Name_Key)
	hash := c.GetString(conf.File_Hash_Key)

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
		log.Error("MakeDirHandler uuid empty")
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
	err := general.UploadObject(&models.UploadObjectParams{
		FileKey:        fileKey,
		User_Uuid:      user_uuid,
		Parent:         user_file_uuid_parent,
		Hash:           hash,
		Size:           conf.Folder_Default_Size,
		File_Uuid:      file_uuid,
		User_File_Uuid: user_file_uuid,
	}, strings.NewReader(""), false)
	if err != nil {
		log.Error("UploadHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_UPLOAD_CODE,
			"msg":  fmt.Sprintf(conf.UPLOAD_FAIL_MESSAGE, fileKey),
		})
		return
	}

	log.Info("UploadHandler success: ", file_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.SUCCESS_RESP_MESSAGE,
		"file_id": user_file_uuid,
		"msg":     fmt.Sprintf(conf.UPLOAD_SUCCESS_MESSAGE),
	})
}
