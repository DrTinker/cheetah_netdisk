package object

import (
	"NetDisk/client"
	"NetDisk/conf"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetFileInfoByPathHandler(c *gin.Context) {
	// 获取路径
	path := c.Query(conf.File_Path_Key)
	if path == "" {
		log.Error("GetFileInfoByPathHandler err: invaild path")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取文件数据
	user_file, err := client.GetDBClient().GetFileByPath(path)
	if err != nil {
		log.Error("GetFileInfoByPathHandler: get user file error ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_GET_INFO_CODE,
			"msg":  conf.GET_INFO_FAIL_MESSAGE,
		})
		return
	}
	// 成功
	log.Info("GetFileInfoByPathHandler: get user file success, path: ", path)
	c.JSON(http.StatusBadRequest, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
		"info": user_file,
	})
}
