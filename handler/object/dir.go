package object

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func MakeDirHandler(c *gin.Context) {
	// 获取文件夹路径
	key := c.GetString(conf.File_Name_Form_Key)
	hash := c.GetString(conf.File_Hash_Key)
	folderName, ext, err := helper.SplitFilePath(key)
	if err != nil {
		log.Error("MakeDirHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	parentIdStr := c.PostForm(conf.File_Parent_Form_Key)
	if key == "" || parentIdStr == "" {
		log.Error("UploadHandler empty file key")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	parentId, _ := strconv.Atoi(parentIdStr)
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
	// 检测用户空间
	now, _, err := volumeCheck(conf.Folder_Default_Size, user_uuid)
	if err != nil {
		log.Error("UploadHandler volume err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_VOLUME_COUNT_CODE,
			"msg":  conf.GET_VOLUME_ERROR_MESSAGE,
		})
		return
	}
	// 存储
	folder_uuid := helper.GenFid(key)
	user_folder_uuid := helper.GenUserFid(user_uuid, key)
	// 拼装结构体
	fileDB := &models.File{
		Uuid: folder_uuid,
		Name: folderName,
		Ext:  ext,
		Path: key,
		Hash: hash,
		Size: conf.Folder_Default_Size,
	}
	userFileDB := &models.UserFile{
		Uuid:      user_folder_uuid,
		User_Uuid: user_uuid,
		Parent_Id: parentId,
		File_Uuid: folder_uuid,
		Name:      folderName,
		Ext:       ext,
	}
	err = client.GetDBClient().CreateUploadRecord(fileDB, userFileDB, now)
	if err != nil {
		log.Error("UploadHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_UPLOAD_CODE,
			"msg":  fmt.Sprintf(conf.UPLOAD_FAIL_MESSAGE, folderName),
		})
		return
	}

	err = client.GetCOSClient().UploadStream(key, strings.NewReader(""))
	if err != nil {
		log.Error("UploadHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_UPLOAD_CODE,
			"msg":  fmt.Sprintf(conf.UPLOAD_FAIL_MESSAGE, folderName),
		})
		return
	}

	log.Info("UploadHandler success: ", folder_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.SUCCESS_RESP_MESSAGE,
		"file_id": user_folder_uuid,
		"msg":     fmt.Sprintf(conf.UPLOAD_SUCCESS_MESSAGE),
	})
}
