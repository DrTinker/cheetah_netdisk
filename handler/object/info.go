package object

import (
	"NetDesk/client"
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 通过文件夹uuid获取该文件下全部文件信息
func GetFileListHandler(c *gin.Context) {
	// 获取传入文件夹uuid
	folder_uuid := c.Query(conf.Folder_Uuid_Key)
	// 获取页号
	pageNumStr := c.Query(conf.Page_Num_Key)
	PageNum, err := strconv.Atoi(pageNumStr)
	if err != nil {
		log.Error("GetFileListHandler get page err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	ext := c.Query(conf.File_Ext_Key)
	// 查询数据库
	// 通过uuid获取ID
	uuids := make([]string, 1)
	uuids[0] = folder_uuid
	ids, err := client.GetDBClient().GetUserFileIDByUuid(uuids)
	if err != nil || ids == nil {
		log.Error("GetFileListHandler get id err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	id := ids[folder_uuid]
	// 查询数据库获取列表
	files, err := client.GetDBClient().GetUserFileListPage(id, PageNum, conf.Default_Page_Size, ext)
	if err != nil || files == nil {
		log.Error("GetFileListHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.LIST_FILES_FAIL_MESSAGE,
		})
		return
	}
	// 处理数据
	show := make([]*models.UserFileShow, len(files))
	for i, file := range files {
		show[i] = &models.UserFileShow{}
		show[i].Uuid = file.Uuid
		show[i].User_Uuid = file.User_Uuid
		show[i].Name = file.Name
		show[i].Ext = file.Ext
		show[i].CreatedAt = helper.TimeFormat(file.CreatedAt)
		show[i].UpdatedAt = helper.TimeFormat(file.UpdatedAt)
	}
	// 返回数据
	log.Info("GetFileListHandler success: ", len(show))
	c.JSON(http.StatusOK, gin.H{
		"code":      conf.HTTP_SUCCESS_CODE,
		"msg":       conf.LIST_FILES_SUCCESS_MESSAGE,
		"file_list": show,
	})
}

func GetFileInfoByPathHandler(c *gin.Context) {
	// 获取路径
	path := c.Query(conf.File_Path_Key)
	if path == "" {
		log.Error("GetFileInfoByPathHandler err: invaild path")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取文件数据
	user_file, err := client.GetDBClient().GetUserFileByPath(path)
	if err != nil {
		log.Error("GetFileInfoByPathHandler: get user file error ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.ERROR_GET_INFO_CODE,
			"msg":  conf.GET_INFO_FAIL_MESSAGE,
		})
		return
	}
	// 成功
	log.Info("GetFileInfoByPathHandler: get user file success, path: ", path)
	c.JSON(http.StatusBadRequest, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
		"info": user_file,
	})
}
