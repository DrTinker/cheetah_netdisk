package trans

import (
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// params: page: 页号
//		   mod: 上传or下载
//		   status: 状态（0传输中，1成功，2失败）
// return: trans_list: 传输列表
// 通过文件夹uuid获取该文件下全部文件信息
func GetTransListHandler(c *gin.Context) {
	// 获取用户ID
	var user_uuid string
	if idstr, f := c.Get(conf.UserID); f {
		user_uuid = helper.Strval(idstr)
	}
	if user_uuid == "" {
		log.Error("GetTransListHandler user id empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取页号
	pageNumStr := c.Query(conf.Page_Num_Key)
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil {
		log.Error("GetTransListHandler get page err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	modStr := c.Query(conf.Trans_Isdown_Key)
	mod, err := strconv.Atoi(modStr)
	if err != nil {
		log.Error("GetTransListHandler mod err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	statusStr := c.Query(conf.Trans_Status_Key)
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
	trans, err := service.GetTransList(user_uuid, pageNum, mod, status)
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
		"code":       conf.HTTP_SUCCESS_CODE,
		"msg":        conf.LIST_TRANS_SUCCESS_MESSAGE,
		"trans_list": trans,
	})
}
