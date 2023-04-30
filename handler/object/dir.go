package object

import (
	"NetDesk/common/conf"
	"NetDesk/common/helper"
	"NetDesk/common/models"
	"NetDesk/service1"

	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 文件系统仅在服务端维护，不在cos存储中体现
func MakeDirHandler(c *gin.Context) {
	// 文件hash值
	hash := c.GetString(conf.File_Hash_Key)
	// 文件夹名称
	folderPath := c.PostForm(conf.File_Name_Key)
	if folderPath == "" {
		log.Error("UploadHandler empty folder name")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}

	user_file_uuid_parent := c.PostForm(conf.Folder_Uuid_Key)
	if folderPath == "" || user_file_uuid_parent == "" {
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
	fileKey := helper.GenFileKey(hash, conf.Folder_Default_EXT)
	user_file_uuid := helper.GenUserFid(user_uuid, fileKey)
	name, _, _ := helper.SplitFilePath(folderPath)
	// 插入数据库记录
	folder := &models.UserFile{
		Uuid:      user_file_uuid,
		User_Uuid: user_uuid,
		Name:      name,
		Ext:       conf.Folder_Default_EXT,
	}
	err := service.Mkdir(folder, user_file_uuid_parent)
	if err != nil {
		log.Error("MakeDirHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_UPLOAD_CODE,
			"msg":  fmt.Sprintf(conf.UPLOAD_FAIL_MESSAGE, fileKey),
		})
		return
	}

	log.Info("MakeDirHandler success: ", user_file_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.SUCCESS_RESP_MESSAGE,
		"file_id": user_file_uuid,
		"msg":     fmt.Sprintf(conf.UPLOAD_SUCCESS_MESSAGE),
	})
}
