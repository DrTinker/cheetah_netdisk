package object

import (
	"NetDisk/conf"
	"NetDisk/helper"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 仅在逻辑复制，COS中不进行实际复制
func CopyHandler(c *gin.Context) {
	// 获取原地址和目的地址
	src := c.PostForm(conf.File_Src_Key)
	des := c.PostForm(conf.File_Des_Key)
	if src == "" || des == "" {
		log.Error("CopyHandler err: empty src or des")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取
	// 获取用户ID
	var user_uuid string
	if idstr, f := c.Get(conf.User_ID); f {
		user_uuid = helper.Strval(idstr)
	}
	if user_uuid == "" {
		log.Error("UploadHandler uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
}
