package user

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

// 返回用户非敏感信息，用于查询分享者信息，需要传入参数，只返回非敏感信息
func UserProfileHandler(c *gin.Context) {
	// 获取用户uuid
	UserUuid := c.Query(conf.UserID)
	if UserUuid == "" {
		log.Error("UserProfileHandler user uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	info, err := client.GetDBClient().GetUserByID(UserUuid)
	if err != nil && err != conf.DBNotFoundError {
		log.Error("LoginHandler pwd err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	res := models.UserInfo{
		Uuid:  UserUuid,
		Name:  info.Name,
		Level: info.Level,
	}
	// 返回成功
	log.Info("UserInfoHandler success: ", UserUuid)
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
		"info": res,
	})
}

// 查询用户自己的信息，返回全部数据,无需传入参数
func UserInfoHandler(c *gin.Context) {
	// 获取用户uuid
	var UserUuid string
	if idstr, f := c.Get(conf.UserID); f {
		UserUuid = helper.Strval(idstr)
	}
	if UserUuid == "" {
		log.Error("UserInfoHandler user uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取空间大小
	info, err := client.GetDBClient().GetUserByID(UserUuid)
	if err != nil && err != conf.DBNotFoundError {
		log.Error("UserInfoHandler pwd err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 返回成功
	log.Info("UserInfoHandler success: ", UserUuid)
	// 返回值去掉密码字段
	info.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
		"info": info,
	})
}
