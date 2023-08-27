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
	// 生成缩略图桶前缀
	cfg, err := client.GetConfigClient().GetCOSConfig()
	if cfg == nil || err != nil {
		log.Error("GetFileListHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.LIST_FILES_FAIL_MESSAGE,
		})
	}
	// 生成结构体
	for i, file := range files {
		tn := ""
		if file.Thumbnail != "" {
			tn = cfg.Domain + "/" + file.Thumbnail
		}
		show[i] = &models.UserFileShow{}
		show[i].Uuid = file.Uuid
		show[i].User_Uuid = file.User_Uuid
		show[i].Name = file.Name
		show[i].Ext = file.Ext
		show[i].Thumbnail = tn
		show[i].Size = file.Size
		show[i].Hash = file.Hash
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

func GetFileInfoHandler(c *gin.Context) {
	// 获取路径
	user_file_uuid := c.Query(conf.File_Uuid_Key)
	if user_file_uuid == "" {
		log.Error("GetFileInfoHandler err: invaild id")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取文件数据
	file, err := client.GetDBClient().GetUserFileByUuid(user_file_uuid)
	if err != nil {
		log.Error("GetFileInfoHandler: get user file error ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.GET_INFO_FAIL_MESSAGE,
		})
		return
	}
	// 生成缩略图桶前缀
	cfg, err := client.GetConfigClient().GetCOSConfig()
	if cfg == nil || err != nil {
		log.Error("GetFileListHandler err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.LIST_FILES_FAIL_MESSAGE,
		})
	}
	tn := ""
	if file.Thumbnail != "" {
		tn = cfg.Domain + "/" + file.Thumbnail
	}
	show := &models.UserFileShow{}
	show.Uuid = file.Uuid
	show.User_Uuid = file.User_Uuid
	show.Name = file.Name
	show.Ext = file.Ext
	show.Thumbnail = tn
	show.Size = file.Size
	show.Hash = file.Hash
	show.CreatedAt = helper.TimeFormat(file.CreatedAt)
	show.UpdatedAt = helper.TimeFormat(file.UpdatedAt)
	// 成功
	log.Info("GetFileInfoHandler: get user file success: ", user_file_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code": conf.HTTP_SUCCESS_CODE,
		"msg":  conf.SUCCESS_RESP_MESSAGE,
		"info": show,
	})
}
