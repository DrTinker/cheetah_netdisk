package share

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/models"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func UpdateShareHandler(c *gin.Context) {
	// 获取share uuid
	shareUuid := c.PostForm(conf.ShareUuid)
	if shareUuid == "" {
		log.Error("UpdateShareHandler share uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取code
	code := c.PostForm(conf.ShareCode)
	if code == "" {
		log.Error("UpdateShareHandler code empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取过期时间并转化
	expireStr := c.PostForm(conf.ShareExpireTime)
	var expire sql.NullTime
	if expireStr != "" {
		tmpExpire, err := time.Parse("2006-01-02 15:04:05", expireStr)
		if err != nil {
			log.Error("UpdateShareHandler ExpireTime invaild")
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.HTTP_INVALID_PARAMS_CODE,
				"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
			})
			return
		}
		expire = sql.NullTime{Time: tmpExpire, Valid: true}
	}
	// 封装结构体
	share := &models.Share{
		Uuid:       shareUuid,
		Code:       code,
		ExpireTime: expire,
	}
	// 调用
	err := client.GetDBClient().UpdateShareByUuid(shareUuid, share)
	if err != nil {
		log.Error("UpdateShareHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 成功
	log.Info("UpdateShareHandler success: ", shareUuid)
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
	})
}
