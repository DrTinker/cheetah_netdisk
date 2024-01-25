package trans

import (
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// params: page: 页号
//
//	mod: 上传or下载
//	status: 状态（0传输中，1成功，2失败）
//
// return: trans_list: 传输列表
// 通过文件夹uuid获取该文件下全部文件信息
func GetTransListHandler(c *gin.Context) {
	// 获取用户ID
	var UserUuid string
	if idstr, f := c.Get(conf.UserID); f {
		UserUuid = helper.Strval(idstr)
	}
	if UserUuid == "" {
		log.Error("GetTransListHandler user id empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取页号
	pageNumStr := c.Query(conf.PageNumKey)
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil {
		log.Error("GetTransListHandler get page err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	modStr := c.Query(conf.TransIsdownKey)
	mod, err := strconv.Atoi(modStr)
	if err != nil {
		log.Error("GetTransListHandler mod err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	statusStr := c.Query(conf.TransStatusKey)
	status, err := strconv.Atoi(statusStr)
	if err != nil {
		log.Error("GetTransListHandler mod err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 查询数据库获取列表
	trans, err := service.GetTransList(UserUuid, pageNum, mod, status)
	if err != nil || trans == nil {
		log.Error("GetTransListHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.GET_TRANS_INFO_FAIL_MESSAGE,
		})
		return
	}

	// 返回数据
	log.Info("GetTransListHandler success: ", len(trans))
	c.JSON(http.StatusOK, gin.H{
		"code":      conf.HTTP_SUCCESS_CODE,
		"msg":       conf.LIST_TransSuccess_MESSAGE,
		"transList": trans,
	})
}
