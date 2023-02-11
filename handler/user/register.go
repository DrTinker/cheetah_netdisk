package user

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"fmt"
	"net/http"

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

	err = client.GetDBClient().CreateUser(&user)
	if err != nil {
		log.Error("RegisterHandler err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.ERROR_REGISTER_CODE,
			"msg":  conf.REGISTER_ERROR_MESSAGE,
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

// TODO 验证码生成
func EmailVerifyHandler(c *gin.Context) {
	// 获取配置文件
	cfg, err := client.GetConfigClient().GetEmailConfig()
	if err != nil {
		log.Error("EmailVerifyHandler err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.ERROR_REGISTER_CODE,
			"msg":  conf.REGISTER_ERROR_MESSAGE,
		})
	}
	// 获取参数
	to := c.Query(conf.User_Email)
	// 生成验证码
	code := helper.GenRandCode()
	err = client.GetCacheClient().SetWithExpire(conf.Code_Cache_Key, code, conf.Code_Expire)
	if err != nil {
		log.Error("EmailVerifyHandler err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.ERROR_VERIFY_CODE,
			"msg":  conf.VERIFY_CODE_GEN_ERROR_MESSAGE,
		})
	}
	// 发送邮件
	content := fmt.Sprintf(conf.Email_Verify_Page, code)
	err = client.GetMsgClient().SendHTMLWithTls(cfg, to, content, conf.Email_Verify_MSG)
	if err != nil {
		log.Error("EmailVerifyHandler err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.ERROR_EMAIL_SEND_CODE,
			"msg":  conf.FAIL_EMAIL_MESSAGE,
		})
		return
	}

	// 返回信息
	log.Info("EmailVerifyHandler: email send success ", to)
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_EMAIL_MESSAGE,
	})
}
