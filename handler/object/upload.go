package object

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"fmt"
	"net/http"
	"strconv"

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
	// TODO 前端传入uuid后端查询id
	parentIdStr := c.PostForm(conf.File_Parent_Form_Key)
	if fileKey == "" || parentIdStr == "" {
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
		log.Error("UploadHandler uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 检测用户空间
	now, total, err := volumeCheck(file.Size, user_uuid)
	if err != nil {
		log.Error("UploadHandler volume err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_VOLUME_COUNT_CODE,
			"msg":  conf.GET_VOLUME_ERROR_MESSAGE,
		})
		return
	}
	// user_file 和 file_pool 插入记录
	// 生成文件uuid
	file_uuid := helper.GenFid(fileKey)
	user_file_uuid := helper.GenUserFid(user_uuid, fileKey)
	// 从文件KEY中获取文件名称
	name, ext, err := helper.SplitFilePath(fileKey)
	if err != nil {
		log.Error("UploadHandler file key err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 拼装结构体
	fileDB := &models.File{
		Uuid: file_uuid,
		Name: name,
		Ext:  ext,
		Path: fileKey,
		Hash: hash,
		Size: int(file.Size),
	}
	userFileDB := &models.UserFile{
		Uuid:      user_file_uuid,
		User_Uuid: user_uuid,
		Parent_Id: parentId,
		File_Uuid: file_uuid,
		Name:      name,
		Ext:       ext,
	}
	// 存储
	err = client.GetDBClient().CreateUploadRecord(fileDB, userFileDB, now)
	if err != nil {
		log.Error("UploadHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_UPLOAD_CODE,
			"msg":  fmt.Sprintf(conf.UPLOAD_FAIL_MESSAGE, file.Filename),
		})
		return
	}
	// 上传cos, 先写数据库再上传
	fd, err := file.Open()
	if err != nil {
		log.Error("UploadHandler file open err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	err = client.GetCOSClient().UploadStream(fileKey, fd)
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
		"file_id": user_file_uuid,
		"now":     total - now - file.Size,
		"msg":     fmt.Sprintf(conf.UPLOAD_SUCCESS_MESSAGE),
	})
}
