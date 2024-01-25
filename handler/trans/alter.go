package trans

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"NetDisk/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 删除已完成或是已失败的记录
func DelTransRecordHandler(c *gin.Context) {
	// 获取transID
	transUuid := c.PostForm(conf.TransUuidKey)
	if transUuid == "" {
		log.Error("DelTransRecordHandler empty trans uuid")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 删除
	err := client.GetDBClient().DelTransByUuid(transUuid)
	if err != nil {
		log.Error("DelTransRecordHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.GET_TRANS_INFO_FAIL_MESSAGE,
		})
		return
	}

	// 返回数据
	log.Info("DelTransRecordHandler success: ", transUuid)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.HTTP_SUCCESS_CODE,
		"msg":     conf.LIST_TransSuccess_MESSAGE,
		"transID": transUuid,
	})
}

// 批量删除
func DelTransBatchHandler(c *gin.Context) {
	listJson, err := c.GetRawData()
	if err != nil {
		log.Error("DelTransBatchHandler get json err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	taskList := &models.BatchTaskInfo{}
	err = json.Unmarshal([]byte(listJson), taskList)
	if err != nil {
		log.Error("DelTransBatchHandler json parse err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 删除
	for _, trans := range taskList.Src {
		err := client.GetDBClient().DelTransByUuid(trans)
		if err != nil {
			log.Error("DelTransRecordHandler err: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": conf.SERVER_ERROR_CODE,
				"msg":  conf.GET_TRANS_INFO_FAIL_MESSAGE,
			})
			return
		}
	}
	// 返回数据
	log.Info("DelTransRecordHandler success: ", len(taskList.Src))
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.LIST_TransSuccess_MESSAGE,
	})
}

// 批量删除
func ClearTransListHandler(c *gin.Context) {
	// 从ctx获取用户ID
	var UserUuid string
	if idstr, f := c.Get(conf.UserID); f {
		UserUuid = helper.Strval(idstr)
	}
	if UserUuid == "" {
		log.Error("ClearTransListHandler user id empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取mod
	modStr := c.PostForm(conf.TransIsdownKey)
	mod, err := strconv.Atoi(modStr)
	if err != nil {
		log.Error("ClearTransListHandler mod err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取status
	statusStr := c.PostForm(conf.TransStatusKey)
	status, err := strconv.Atoi(statusStr)
	if err != nil {
		log.Error("ClearTransListHandler mod err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 删除
	err = client.GetDBClient().DelTransByStatus(UserUuid, mod, status)
	if err != nil {
		log.Error("ClearTransListHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 返回数据
	log.Info("ClearTransListHandler success: ", UserUuid, " mod: ", mod, " status: ", status)
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.LIST_TransSuccess_MESSAGE,
	})
}

// 取消正在进行的上传
func CancelUploadHandler(c *gin.Context) {
	// 获取transID
	transUuid := c.PostForm(conf.TransUuidKey)
	if transUuid == "" {
		log.Error("CancelUploadHandler empty trans uuid")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 调用service
	err := service.CancelUpload(transUuid)
	if err != nil {
		log.Error("CancelUploadHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  "Cancel upload error",
		})
		return
	}
	// 返回数据
	log.Info("CancelUploadHandler success: ", transUuid)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.HTTP_SUCCESS_CODE,
		"msg":     "Cancel success",
		"transID": transUuid,
	})
}

// 取消正在进行的下载
func CancelDownloadHandler(c *gin.Context) {
	// 获取transID
	transUuid := c.PostForm(conf.TransUuidKey)
	if transUuid == "" {
		log.Error("CancelDownloadHandler empty trans uuid")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 调用service
	err := service.CancelDownload(transUuid)
	if err != nil {
		log.Error("CancelDownloadHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  "Cancel upload error",
		})
		return
	}
	// 返回数据
	log.Info("CancelDownloadHandler success: ", transUuid)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.HTTP_SUCCESS_CODE,
		"msg":     "Cancel success",
		"transID": transUuid,
	})
}
