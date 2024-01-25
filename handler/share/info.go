package share

import (
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 查询分享链接
func GetShareInfoHandler(c *gin.Context) {
	// 获取uuid
	shareUuid := c.Query(conf.ShareUuid)
	if shareUuid == "" {
		log.Error("GetShareInfoHandler share uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 查询数据库
	info, time_out, err := service.GetShareInfo(shareUuid)
	if err == conf.DBNotFoundError || err == conf.FileDeletedError {
		log.Warn("GetShareInfoHandler record not found", shareUuid)
		c.JSON(http.StatusOK, gin.H{
			"code": conf.RECORD_DELETED_CODE,
			"msg":  conf.RECORD_DELETED_MSG,
		})
		return
	}
	if err != nil {
		log.Error("GetShareInfoHandler get info error ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 成功
	log.Info("GetShareInfoHandler success: ", shareUuid)
	// 分享过期
	if time_out {
		c.JSON(http.StatusOK, gin.H{
			"code": conf.WARN_SHARE_EXPIRES_CODE,
			"msg":  conf.SHARE_EXPIRES_MSG,
		})
		return
	}
	// 未过期
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
		"info": info,
	})
}

// 获取分享列表
func GetShareListHandler(c *gin.Context) {
	// 获取用户ID
	var userUuid string
	if idstr, f := c.Get(conf.UserID); f {
		userUuid = helper.Strval(idstr)
	}
	if userUuid == "" {
		log.Error("GetTransListHandler user id empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取页号
	pageNumStr := c.Query(conf.PageNumKey)
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil {
		log.Error("GetShareListHandler get page err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	modStr := c.Query(conf.ShareMod)
	mod, err := strconv.Atoi(modStr)
	if err != nil {
		log.Error("GetShareListHandler mod err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 查询数据库
	info, err := service.GetShareList(userUuid, pageNum, mod)
	if err != nil {
		log.Error("GetShareListHandler get info error ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 成功
	log.Info("GetShareListHandler success: ", userUuid)
	// 未过期
	c.JSON(http.StatusOK, gin.H{
		"code":      conf.HTTP_SUCCESS_CODE,
		"msg":       conf.SUCCESS_RESP_MESSAGE,
		"shareList": info,
	})
}
