package object

import (
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"NetDisk/service"

	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 文件系统仅在服务端维护，不在cos存储中体现
func MakeDirHandler(c *gin.Context) {
	// 文件夹名称
	folderName := c.PostForm(conf.FileNameKey)
	if folderName == "" {
		log.Error("UploadHandler empty folder name")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}

	userFileUuidParent := c.PostForm(conf.FileParentKey)
	if userFileUuidParent == "" {
		log.Error("UploadHandler empty file key")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}

	// 获取用户ID
	var UserUuid string
	if idstr, f := c.Get(conf.UserID); f {
		UserUuid = helper.Strval(idstr)
	}
	if UserUuid == "" {
		log.Error("MakeDirHandler uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 生成ID
	userFileUuid := helper.GenUserFid(UserUuid, folderName)
	// 插入数据库记录
	folder := &models.UserFile{
		Uuid:     userFileUuid,
		UserUuid: UserUuid,
		Name:     folderName,
		Ext:      conf.FolderDefaultExt,
	}
	err := service.Mkdir(folder, userFileUuidParent)
	if err != nil {
		log.Error("MakeDirHandler err: ", err)
		c.JSON(http.StatusOK, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}

	log.Info("MakeDirHandler success: ", userFileUuid)
	c.JSON(http.StatusOK, gin.H{
		"code":   conf.HTTP_SUCCESS_CODE,
		"fileID": userFileUuid,
		"msg":    fmt.Sprintf(conf.UPLOAD_SUCCESS_MESSAGE),
	})
}
