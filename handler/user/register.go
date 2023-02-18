package user

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/handler/general"
	"NetDisk/helper"
	"NetDisk/models"
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(c *gin.Context) {
	// 初始化user struct
	user := models.User{}
	err := c.ShouldBind(&user)
	if err != nil {
		log.Error("RegisterHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 判断是否存在用户
	info, err := client.GetDBClient().GetUserByEmail(user.Email)
	if err != nil && err != conf.DBNotFoundError {
		log.Error("RegisterHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_REGISTER_CODE,
			"msg":  conf.REGISTER_ERROR_MESSAGE,
		})
	}
	// 存在则报错
	if info != nil {
		log.Info("RegisterHandler: repeat", info.Email)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_REGISTER_CODE,
			"msg":  conf.REGISTER_REPEAT_MESSAGE,
		})
		return
	}

	// 判断验证码是否有效
	src := c.Query(conf.Code_Param_Key)
	code, err := client.GetCacheClient().Get(conf.Code_Cache_Key)
	if err != nil || code == "" || src != code {
		log.Error("RegisterHandler: verify code error ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_REGISTER_CODE,
			"msg":  conf.VERIFY_CODE_ERROR_MESSAGE,
		})
		return
	}
	// 生成用户ID
	id := helper.GenUid(user.Name, user.Email)

	user.Uuid = id
	user.Level = conf.User_Level_normal
	user.Now_Volume = 0
	user.Total_Volume = conf.User_Normal_Volume // 单位B

	// 生成用户空间根目录uuid
	fileKey := fmt.Sprintf("%s/%s-%s/", conf.Default_System_Prefix, user.Name, user.Uuid)
	file_uuid := helper.GenFid(fileKey)
	user_file_uuid := helper.GenUserFid(user.Uuid, fileKey)
	user.Start_Uuid = user_file_uuid

	// 创建用户记录
	err = client.GetDBClient().CreateUser(&user)
	if err != nil {
		log.Error("RegisterHandler err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.ERROR_REGISTER_CODE,
			"msg":  conf.REGISTER_ERROR_MESSAGE,
		})
		return
	}

	// 创建用户文件空间根目录
	// 上传
	err = general.UploadObjectServer(&models.UploadObjectParams{
		FileKey:        fileKey,
		User_Uuid:      user.Uuid,
		Hash:           fmt.Sprintf("%x", md5.Sum([]byte(fileKey))),
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
	// 返回成功
	log.Info("RegisterHandler success: ", user.Uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.HTTP_SUCCESS_CODE,
		"msg":     conf.SUCCESS_RESP_MESSAGE,
		"user_id": id,
	})
}
