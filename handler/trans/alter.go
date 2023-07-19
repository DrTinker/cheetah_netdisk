package trans

import (
	"NetDesk/client"
	"NetDesk/conf"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func DelTransRecordHandler(c *gin.Context) {
	// 获取页号
	trans_uuid := c.PostForm(conf.Trans_Uuid_Key)
	if trans_uuid == "" {
		log.Error("DelTransRecordHandler empty trans uuid")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 查询数据库获取列表
	err := client.GetDBClient().DelTransByUuid(trans_uuid)
	if err != nil {
		log.Error("DelTransRecordHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.GET_TRANS_INFO_FAIL_MESSAGE,
		})
		return
	}

	// 返回数据
	log.Info("DelTransRecordHandler success: ", trans_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":     conf.HTTP_SUCCESS_CODE,
		"msg":      conf.LIST_TRANS_SUCCESS_MESSAGE,
		"trans_id": trans_uuid,
	})
}
