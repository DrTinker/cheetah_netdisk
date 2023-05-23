package share

import (
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
	"NetDesk/service"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func CopyFileByShareHandler(c *gin.Context) {
	// 获取用户uuid
	var user_uuid string
	if idstr, f := c.Get(conf.User_ID); f {
		user_uuid = helper.Strval(idstr)
	}
	if user_uuid == "" {
		log.Error("CopyFileByShareHandler user uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取des_uuid
	share_uuid := c.PostForm(conf.Share_Uuid)
	des_uuid := c.PostForm(conf.File_Des_Key)
	if share_uuid == "" || des_uuid == "" {
		log.Error("CopyFileByShareHandler err: empty share_uuid or des_uuid")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 调用service
	_, expires, err := service.GetShareInfo(share_uuid)
	if err == conf.DBNotFoundError {
		log.Warn("GetShareInfoHandler record not found", share_uuid)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.RECORD_DELETED_MSG,
		})
		return
	}
	if err != nil {
		log.Error("CopyFileByShareHandler get share info ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 过期
	if expires {
		log.Error("CopyFileByShareHandler share expires ")
		c.JSON(http.StatusOK, gin.H{
			"code": conf.WARN_SHARE_EXPIRES_CODE,
			"msg":  conf.SHARE_EXPIRES_MSG,
		})
		return
	}
	err = service.CopyFileByShare(share_uuid, des_uuid, user_uuid)
	if err != nil {
		log.Error("CopyFileByShareHandler copy file err ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 成功
	log.Info("CopyFileByShareHandler success: ", share_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
	})
}

func CopyFileByShareBatchHandler(c *gin.Context) {
	// 获取用户uuid
	var user_uuid string
	if idstr, f := c.Get(conf.User_ID); f {
		user_uuid = helper.Strval(idstr)
	}
	if user_uuid == "" {
		log.Error("CopyFileByShareBatchHandler user uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取文件uuid user_file
	listJson, err := c.GetRawData()
	if err != nil {
		log.Error("CopyFileByShareBatchHandler get json err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	taskList := &models.BatchTaskInfo{}
	err = json.Unmarshal([]byte(listJson), taskList)
	if err != nil {
		log.Error("CopyFileByShareBatchHandler json parse err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// share uuid
	share_uuid := c.PostForm(conf.Share_Uuid)
	// 调用service
	_, expires, err := service.GetShareInfo(share_uuid)
	if err == conf.DBNotFoundError {
		log.Warn("CopyFileByShareBatchHandler record not found", share_uuid)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.RECORD_DELETED_MSG,
		})
		return
	}
	if err != nil {
		log.Error("CopyFileByShareBatchHandler get share info ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 过期
	if expires {
		log.Error("CopyFileByShareBatchHandler share expires ")
		c.JSON(http.StatusOK, gin.H{
			"code": conf.WARN_SHARE_EXPIRES_CODE,
			"msg":  conf.SHARE_EXPIRES_MSG,
		})
		return
	}
	// 复制
	success := make([]string, 0)
	fail := make([]string, 0)
	for _, task := range taskList.Src {
		err := service.CopyObject(task, taskList.Des, user_uuid)
		if err != nil {
			log.Error("CopyFileByShareBatchHandler copy err: ", err)
			fail = append(fail, task)
		} else {
			success = append(success, task)
		}
	}

	log.Info("CopyFileByShareBatchHandler success: ", taskList)
	c.JSON(http.StatusBadRequest, gin.H{
		"code":    conf.HTTP_SUCCESS_CODE,
		"msg":     conf.SUCCESS_RESP_MESSAGE,
		"success": success,
		"fail":    fail,
		"total":   len(taskList.Src),
	})
}
