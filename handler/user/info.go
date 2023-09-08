package user

import (
	"NetDesk/client"
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

// 返回用户非敏感信息，用于查询分享者信息，需要传入参数，只返回非敏感信息
func UserProfileHandler(c *gin.Context) {
	// 获取用户uuid
	user_uuid := c.Query(conf.UserID)
	if user_uuid == "" {
		log.Error("UserProfileHandler user uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	info, err := client.GetDBClient().GetUserByID(user_uuid)
	if err != nil && err != conf.DBNotFoundError {
		log.Error("LoginHandler pwd err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	res := models.UserInfo{
		Uuid:  user_uuid,
		Name:  info.Name,
		Level: info.Level,
	}
	// 返回成功
	log.Info("UserInfoHandler success: %v", user_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
		"info": res,
	})
}

// 查询用户自己的信息，返回全部数据,无需传入参数
func UserInfoHandler(c *gin.Context) {
	// 获取用户uuid
	var user_uuid string
	if idstr, f := c.Get(conf.UserID); f {
		user_uuid = helper.Strval(idstr)
	}
	if user_uuid == "" {
		log.Error("UserInfoHandler user uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取空间大小
	info, err := client.GetDBClient().GetUserByID(user_uuid)
	if err != nil && err != conf.DBNotFoundError {
		log.Error("UserInfoHandler pwd err: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 返回成功
	log.Info("UserInfoHandler success: %v", user_uuid)
	// 返回值去掉密码字段
	info.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
		"info": info,
	})
}
