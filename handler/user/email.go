package user

import (
	"NetDesk/common/client"
	"NetDesk/common/conf"
	"NetDesk/common/helper"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

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
	// TODO 修改redis key
	// TODO 判断邮箱是否已经被注册
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
