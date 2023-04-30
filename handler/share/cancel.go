package share

import (
	"NetDesk/common/conf"
	"NetDesk/service_old"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 取消分享链接
func CancelShareHandler(c *gin.Context) {
	// 获取share uuid
	share_uuid := c.PostForm(conf.Share_Uuid)
	if share_uuid == "" {
		log.Error("CancelShareHandler share uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 调用service层
	err := service.CancelShare(share_uuid)
	if err != nil {
		log.Error("CancelShareHandler delete share err ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	// 成功
	log.Info("CancelShareHandler success: ", share_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":       conf.HTTP_SUCCESS_CODE,
		"msg":        conf.SUCCESS_RESP_MESSAGE,
		"share_uuid": share_uuid,
	})
}
