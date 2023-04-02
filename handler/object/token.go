package object

import (
	"NetDisk/client"
	"NetDisk/conf"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetTokenHandler(c *gin.Context) {
	// 获取文件路径
	fileKey := c.Query(conf.File_Path_Key)
	if fileKey == "" {
		log.Error("GetTokenHandler err: invaild file path")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取预签名
	url, err := client.GetCOSClient().GetPresignedUrl(fileKey, conf.Default_Sign_Expire)
	if err != nil {
		log.Error("GetTokenHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_GET_URL_CODE,
			"msg":  conf.GET_SIGN_ERROR_MESSAGE,
		})
		return
	}
	log.Info("GetTokenHandler success: ", fileKey)
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
		"sign": url,
	})
}