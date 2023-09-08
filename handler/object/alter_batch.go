package object

import (
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
	"NetDesk/service"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// TODO 抽象为task统一操作
// 仅在逻辑复制，COS中不进行实际复制
func CopyFileBatchHandler(c *gin.Context) {
	// 获取原地址和目的地址列表
	listJson, err := c.GetRawData()
	if err != nil {
		logrus.Error("CopyFileBatchHandler get json err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	taskList := &models.BatchTaskInfo{}
	err = json.Unmarshal([]byte(listJson), taskList)
	if err != nil {
		logrus.Error("CopyFileBatchHandler json parse err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取用户ID
	var user_uuid string
	if idstr, f := c.Get(conf.UserID); f {
		user_uuid = helper.Strval(idstr)
	}
	if user_uuid == "" {
		logrus.Error("CopyFileBatchHandler uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 复制
	success := make([]string, 0)
	fail := make([]string, 0)
	for _, task := range taskList.Src {
		err := service.CopyObject(task, taskList.Des, user_uuid)
		if err != nil {
			logrus.Error("CopyFileBatchHandler copy err: ", err)
			fail = append(fail, task)
		} else {
			success = append(success, task)
		}
	}
	logrus.Info("CopyFileBatchHandler copy batch success: ", taskList)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.HTTP_SUCCESS_CODE,
		"msg":     conf.SUCCESS_RESP_MESSAGE,
		"success": success,
		"fail":    fail,
		"total":   len(taskList.Src),
	})
}

func MoveFileBatchHandler(c *gin.Context) {
	// 获取原地址和目的地址列表
	listJson, err := c.GetRawData()
	if err != nil {
		logrus.Error("MoveFileBatchHandler get json err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	taskList := &models.BatchTaskInfo{}
	err = json.Unmarshal([]byte(listJson), taskList)
	if err != nil {
		logrus.Error("MoveFileBatchHandler json parse err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 复制
	success := make([]string, 0)
	fail := make([]string, 0)
	for _, task := range taskList.Src {
		err := service.MoveObject(task, taskList.Des)
		if err != nil {
			logrus.Error("MoveFileBatchHandler copy err: ", err)
			fail = append(fail, task)
		} else {
			success = append(success, task)
		}
	}
	logrus.Info("MoveFileBatchHandler copy batch success: ", taskList)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.HTTP_SUCCESS_CODE,
		"msg":     conf.SUCCESS_RESP_MESSAGE,
		"success": success,
		"fail":    fail,
		"total":   len(taskList.Src),
	})
}

func DeleteFileBatchHandler(c *gin.Context) {
	listJson, err := c.GetRawData()
	if err != nil {
		logrus.Error("DeleteFileBatchHandler get json err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	taskList := &models.BatchTaskInfo{}
	err = json.Unmarshal([]byte(listJson), taskList)
	if err != nil {
		logrus.Error("DeleteFileBatchHandler json parse err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 删除数据库记录
	success := make([]string, 0)
	fail := make([]string, 0)
	for _, task := range taskList.Src {
		err := service.DeleteObject(task)
		if err != nil {
			logrus.Error("DeleteFileBatchHandler delete err: ", err)
			fail = append(fail, task)
		} else {
			success = append(success, task)
		}
	}
	// 成功
	logrus.Info("FileDeleteHandler success: ", taskList)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.HTTP_SUCCESS_CODE,
		"msg":     conf.SUCCESS_RESP_MESSAGE,
		"success": success,
		"fail":    fail,
		"total":   len(taskList.Src),
	})
}
