package share

import (
	"NetDesk/common/conf"
	"NetDesk/common/helper"
	"NetDesk/common/models"
	"NetDesk/service1"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

// 创建分享链接
func CreateShareHandler(c *gin.Context) {
	// 获取用户uuid
	var user_uuid string
	if idstr, f := c.Get(conf.User_ID); f {
		user_uuid = helper.Strval(idstr)
	}
	if user_uuid == "" {
		log.Error("CreateShareHandler user uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取user_file_uuid
	user_file_uuid := c.PostForm(conf.Share_User_File_Uuid)
	if user_file_uuid == "" {
		log.Error("CreateShareHandler user file uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取code
	code := c.PostForm(conf.Share_Code)
	if code == "" {
		log.Error("CreateShareHandler code empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取过期时间并转化
	expireStr := c.PostForm(conf.Share_Expire_Time)
	expire, err := time.Parse("2006-01-02 15:04:05", expireStr)
	if err != nil {
		log.Error("CreateShareHandler expire_time invaild")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 生成share uuid
	share_uuid := helper.GenSid(user_uuid, code)
	// 封装结构体
	param := &models.CreateShareParams{
		Share_Uuid:     share_uuid,
		User_Uuid:      user_uuid,
		User_File_Uuid: user_file_uuid,
		Code:           code,
		Expire:         expire,
	}
	// 调用service层
	err = service.CreateShareLink(param)
	if err != nil {
		log.Error("CreateShareHandler create share record err ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_CREATE_SHARE_CODE,
			"msg":  conf.CREATE_SHARE_FAIL_MESSAGE,
		})
		return
	}
	// 成功
	log.Info("CreateShareHandler success: ", share_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":     conf.SUCCESS_RESP_MESSAGE,
		"msg":      conf.SUCCESS_RESP_MESSAGE,
		"share_id": share_uuid,
	})
}
