package object

import (
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/service"
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
	err := service.CopyObject(src, des)
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
	err := service.MoveObject(src, des)
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

// 仅涉及user_file表，不涉及cos
func FileUpdateHandler(c *gin.Context) {
	// 获取文件uuid user_file
	user_file_uuid := c.PostForm(conf.File_Uuid_Key)
	if user_file_uuid == "" {
		log.Error("FileUpdateHandler err: invaild file uuid")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取更改后名称，只传入全名 name.ext
	fullName := c.PostForm(conf.File_Name_Key)
	name, ext, err := helper.SplitFileFullName(fullName)
	if err != nil || name == "" || ext == "" {
		log.Error("FileUpdateHandler err: invaild file name")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 仅更改名称
	if err := service.UpdateObjectName(user_file_uuid, name, ext); err != nil {
		log.Error("FileUpdateHandler update err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_UPDATE_NAME_CODE,
			"msg":  conf.UPDATE_FILE_NAME_FAIL_MESSAGE,
		})
		return
	}
	// 成功
	log.Info("FileUpdateHandler success: ", user_file_uuid)
	c.JSON(http.StatusBadRequest, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
	})
}

func FileDeleteHandler(c *gin.Context) {
	// 获取文件uuid user_file
	user_file_uuid := c.PostForm(conf.File_Uuid_Key)
	if user_file_uuid == "" {
		log.Error("FileDeleteHandler err: invaild file uuid")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 删除数据库记录
	// 判断file_pool中引用数，若未0则删除COS中文件
	if err := service.DeleteObject(user_file_uuid); err != nil {
		log.Error("FileDeleteHandler err: invaild file uuid")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_DELETE_FILE_CODE,
			"msg":  conf.DELETE_FILE_FAIL_MESSAGE,
		})
		return
	}
	// 成功
	log.Info("FileDeleteHandler success: ", user_file_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
	})
}
