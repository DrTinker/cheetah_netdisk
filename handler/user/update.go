package user

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func UpdateUserNameHandler(c *gin.Context) {
	// 获取username
	userName := c.PostForm(conf.UserName)
	if userName == "" {
		log.Error("UpdateUserNameHandler user name empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 检测登录态
	var UserUuid string
	if idstr, f := c.Get(conf.UserID); f {
		UserUuid = helper.Strval(idstr)
	}
	if UserUuid == "" {
		log.Error("UpdateUserNameHandler user uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 更改
	err := client.GetDBClient().UpdateUserName(UserUuid, userName)
	if err != nil {
		log.Error("UpdateUserNameHandler server err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 成功
	log.Info("UpdateUserNameHandler success: ", UserUuid)
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
	})
}
