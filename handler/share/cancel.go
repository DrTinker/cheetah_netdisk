package share

import (
	"NetDisk/conf"
	"NetDisk/models"
	"NetDisk/service"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// 取消分享链接
func CancelShareHandler(c *gin.Context) {
	// 获取share uuid
	ShareUuid := c.PostForm(conf.ShareUuid)
	if ShareUuid == "" {
		log.Error("CancelShareHandler share uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 调用service层
	err := service.CancelShare(ShareUuid)
	if err != nil {
		log.Error("CancelShareHandler delete share err ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 成功
	log.Info("CancelShareHandler success: ", ShareUuid)
	c.JSON(http.StatusOK, gin.H{
		"code":      conf.HTTP_SUCCESS_CODE,
		"msg":       conf.SUCCESS_RESP_MESSAGE,
		"ShareUuid": ShareUuid,
	})
}

// 取消分享链接
func CancelShareBatchHandler(c *gin.Context) {
	// 获取原地址和目的地址列表
	listJson, err := c.GetRawData()
	if err != nil {
		logrus.Error("CancelShareBatchHandler get json err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	taskList := &models.BatchTaskInfo{}
	err = json.Unmarshal([]byte(listJson), &taskList)
	if err != nil {
		logrus.Error("CancelShareBatchHandler json parse err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	cancelList := taskList.Src
	// 调用service层
	err = service.CancelBatchShare(cancelList)
	if err != nil {
		log.Error("CancelShareBatchHandler delete share err ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 成功
	log.Info("CancelShareBatchHandler success: ")
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.HTTP_SUCCESS_CODE,
		"msg":     conf.SUCCESS_RESP_MESSAGE,
		"success": len(cancelList),
	})
}
