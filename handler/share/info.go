package share

import (
	"NetDesk/conf"
	"NetDesk/service"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 查询分享链接
func GetShareInfoHandler(c *gin.Context) {
	// 获取uuid
	share_uuid := c.Query(conf.Share_Uuid)
	if share_uuid == "" {
		log.Error("GetShareInfoHandler share uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 查询数据库
	info, time_out, err := service.GetShareInfo(share_uuid)
	if err == conf.DBNotFoundError {
		log.Warn("GetShareInfoHandler record not found", share_uuid)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
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
	log.Info("GetShareInfoHandler success: ", share_uuid)
	// 分享过期
	if time_out {
		c.JSON(http.StatusOK, gin.H{
			"code": conf.WARN_SHARE_EXPIRES_CODE,
			"msg":  conf.SHARE_EXPIRES_MSG,
			"info": info,
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
