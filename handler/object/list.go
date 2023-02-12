package object

import (
	"NetDisk/client"
	"NetDisk/conf"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 通过文件夹uuid获取该文件下全部文件信息
func GetFileListHandler(c *gin.Context) {
	// 获取传入文件夹uuid
	folder_uuid := c.Query(conf.Folder_Uuid_Key)
	// 查询数据库
	// 通过uuid获取ID
	uuids := make([]string, 1)
	uuids[0] = folder_uuid
	ids, err := client.GetDBClient().GetFileIDByUuid(uuids)
	if err != nil || ids == nil {
		log.Error("GetFileListHandler get id err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	id := ids[0]
	// 查询数据库获取列表
	files, err := client.GetDBClient().GetFileList(id)
	if err != nil || files == nil {
		log.Error("GetFileListHandler get id err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_LIST_FILES_CODE,
			"msg":  conf.LIST_FILES_FAIL_MESSAGE,
		})
		return
	}
	// 返回数据
	log.Info("GetFileListHandler success: ", folder_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":      conf.SUCCESS_RESP_MESSAGE,
		"msg":       conf.LIST_FILES_SUCCESS_MESSAGE,
		"file_list": files,
	})
}
