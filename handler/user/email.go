package user

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
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
		log.Error("EmailVerifyHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.ERROR_REGISTER_CODE,
			"msg":  conf.REGISTER_ERROR_MESSAGE,
		})
	}
	// 获取参数
	to := c.Query(conf.UserEmail)
	// 生成验证码
	code := helper.GenRandCode()
	// 生成rediskey
	key := helper.GenVerifyCodeKey(conf.CodeCacheKey, to)
	// 上一个验证码过期后才能set
	err = client.GetCacheClient().SetWithExpire(key, code, conf.CodeExpire)
	if err != nil {
		log.Error("EmailVerifyHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.ERROR_VERIFY_CODE,
			"msg":  conf.VERIFY_CODE_GEN_ERROR_MESSAGE,
		})
		return
	}
	// 发送邮件
	content := fmt.Sprintf(conf.EmailVerifyPage, code)
	err = client.GetMsgClient().SendHTMLWithTls(cfg, to, content, conf.EmailVerifyMsg)
	if err != nil {
		log.Error("EmailVerifyHandler err: ", err)
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

func ForgetPwdHandler(c *gin.Context) {
	// 获取配置文件
	cfg, err := client.GetConfigClient().GetEmailConfig()
	if err != nil {
		log.Error("ForgetPwdHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
	}
	// 获取邮箱 & 手机号
	email := c.Query(conf.UserEmail)
	if email == "" {
		log.Error("ForgetPwdHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.ERROR_INVAILD_PAGE_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
	}
	phone := c.Query(conf.UserPhone)
	// 查询数据库
	user, err := client.GetDBClient().GetUserByEmail(email)
	if err != nil {
		log.Error("ForgetPwdHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
	}
	// 判断
	if user.Phone == phone {
		// 相同则发送邮件
		content := fmt.Sprintf(conf.ForgetPasswordPage, user.Password)
		err = client.GetMsgClient().SendHTMLWithTls(cfg, email, content, conf.ForgetPasswordMsg)
		if err != nil {
			log.Error("ForgetPwdHandler err: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": conf.SERVER_ERROR_CODE,
				"msg":  conf.SERVER_ERROR_MSG,
			})
			return
		}
	}
	// 返回信息
	log.Info("EmailVerifyHandler: email send success ", email)
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_EMAIL_MESSAGE,
	})
}
