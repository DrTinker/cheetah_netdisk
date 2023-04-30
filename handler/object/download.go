package object

import (
	"NetDesk/common/client"
	"NetDesk/common/conf"
	"NetDesk/common/helper"
	service "NetDesk/service1"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

// 客户端从服务端下载完整文件
func DownloadFileHandler(c *gin.Context) {
	// 获取user_file_uuid
	user_file_uuid := c.Query(conf.File_Uuid_Key)
	if user_file_uuid == "" {
		log.Error("DownloadFileHandler err: user file uuid rmpty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 调用service下载COS文件至tmp
	filePath, err := service.DownloadToTmp(user_file_uuid)
	if filePath == "" || err != nil {
		log.Error("DownloadFileHandler download err ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 查询文件名称
	user_file, err := client.GetDBClient().GetUserFileByUuid(user_file_uuid)
	if err != nil {
		log.Error("DownloadFileHandler get file info err ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 向客户端传递
	// 打开文件
	fileTmp, err := helper.OpenFile(filePath)
	if err != nil {
		log.Error("DownloadFileHandler open file err ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	defer fileTmp.Close()

	// 写入文件至body
	fileName := user_file.Name + "." + user_file.Ext
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")

	c.File(filePath)
}
