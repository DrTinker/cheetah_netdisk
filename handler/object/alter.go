package object

import (
	"NetDisk/conf"
	"NetDisk/handler/general"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 仅在逻辑复制，COS中不进行实际复制
func CopyFileHandler(c *gin.Context) {
	// 获取原地址和目的地址
	src := c.PostForm(conf.File_Src_Key)
	des := c.PostForm(conf.File_Des_Key)
	if src == "" || des == "" {
		log.Error("CopyHandler err: empty src or des")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 复制
	err := general.AlterObject(src, des, 0)
	if err != nil {
		log.Error("CopyHandler copy err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_FILE_COPY_CODE,
			"msg":  conf.COPY_FILE_FAIL_MESSAGE,
		})
		return
	}
	log.Info("CopyHandler copy success: ", src)
	c.JSON(http.StatusBadRequest, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
	})
}

// 移动文件-
func MoveFileHandler(c *gin.Context) {
	// 获取原地址和目的地址
	src := c.PostForm(conf.File_Src_Key)
	des := c.PostForm(conf.File_Des_Key)
	if src == "" || des == "" {
		log.Error("MoveFileHandler err: empty src or des")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 移动
	err := general.AlterObject(src, des, 1)
	if err != nil {
		log.Error("MoveFileHandler copy err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_FILE_COPY_CODE,
			"msg":  conf.COPY_FILE_FAIL_MESSAGE,
		})
		return
	}
	log.Info("MoveFileHandler copy success: ", src)
	c.JSON(http.StatusBadRequest, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
	})
}
