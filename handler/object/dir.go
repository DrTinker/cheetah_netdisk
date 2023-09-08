package object

import (
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
	"NetDesk/service"

	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 文件系统仅在服务端维护，不在cos存储中体现
func MakeDirHandler(c *gin.Context) {
	// 文件夹名称
	folderName := c.PostForm(conf.File_Name_Key)
	if folderName == "" {
		log.Error("UploadHandler empty folder name")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}

	user_file_uuid_parent := c.PostForm(conf.Folder_Uuid_Key)
	if user_file_uuid_parent == "" {
		log.Error("UploadHandler empty file key")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}

	// 获取用户ID
	var user_uuid string
	if idstr, f := c.Get(conf.UserID); f {
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
	user_file_uuid := helper.GenUserFid(user_uuid, folderName)
	// 插入数据库记录
	folder := &models.UserFile{
		Uuid:      user_file_uuid,
		User_Uuid: user_uuid,
		Name:      folderName,
		Ext:       conf.Folder_Default_EXT,
	}
	err := service.Mkdir(folder, user_file_uuid_parent)
	if err != nil {
		log.Error("MakeDirHandler err: ", err)
		c.JSON(http.StatusOK, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}

	log.Info("MakeDirHandler success: ", user_file_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.HTTP_SUCCESS_CODE,
		"file_id": user_file_uuid,
		"msg":     fmt.Sprintf(conf.UPLOAD_SUCCESS_MESSAGE),
	})
}
