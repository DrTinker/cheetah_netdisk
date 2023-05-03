package handler

import (
	"NetDesk/common/conf"
	"NetDesk/service/apigw/logic"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

// TODO 验证码生成
func SendSignUpEmailHandler(c *gin.Context) {
	// 获取参数
	to := c.Query(conf.User_Email)
	// 调用logic
	l, err := logic.NewNoticeLogic()
	if err != nil {
		log.Error("SendSignUpEmailHandler: get logic err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	resp, err := l.SendSignUpEmailLogic(to)
	if err != nil {
		log.Error("SendSignUpEmailHandler: send email err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}

	// 返回信息
	log.Info("EmailVerifyHandler: email send success ", to)
	c.JSON(http.StatusOK, gin.H{
		"code": resp.Resp.Code,
		"msg":  resp.Resp.RespMsg,
	})
}
