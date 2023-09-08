package share

import (
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
	"NetDesk/service"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

// 创建分享链接
func SetShareHandler(c *gin.Context) {
	// 获取用户uuid
	var user_uuid string
	if idstr, f := c.Get(conf.UserID); f {
		user_uuid = helper.Strval(idstr)
	}
	if user_uuid == "" {
		log.Error("SetShareHandler user uuid empty")
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
	// 获取code
	fullName := c.PostForm(conf.Share_Name)
	if fullName == "" {
		log.Error("CreateShareHandler name empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取过期时间并转化
	expireStr := c.PostForm(conf.Share_Expire_Time)
	var expire sql.NullTime
	if expireStr != "" {
		tmpExpire, err := time.Parse("2006-01-02 15:04:05", expireStr)
		if err != nil {
			log.Error("CreateShareHandler expire_time invaild")
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.HTTP_INVALID_PARAMS_CODE,
				"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
			})
			return
		}
		expire = sql.NullTime{Time: tmpExpire, Valid: true}
	}

	var share_uuid string
	// 看是否传入
	share_uuid = c.PostForm(conf.Share_Uuid)
	if share_uuid == "" {
		// 没传入则生成生成share uuid
		share_uuid = helper.GenSid(user_uuid, code)
	}

	// 封装结构体
	param := &models.CreateShareParams{
		Share_Uuid:     share_uuid,
		User_Uuid:      user_uuid,
		User_File_Uuid: user_file_uuid,
		Fullname:       fullName,
		Code:           code,
		Expire:         expire,
	}
	// 调用service层
	err := service.SetShareLink(param)
	if err != nil {
		log.Error("CreateShareHandler create share record err ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.CREATE_SHARE_FAIL_MESSAGE,
		})
		return
	}
	// 成功
	log.Info("CreateShareHandler success: ", share_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":     conf.HTTP_SUCCESS_CODE,
		"msg":      conf.SUCCESS_RESP_MESSAGE,
		"share_id": share_uuid,
	})
}

func CreateShareBatchHandler(c *gin.Context) {
	// 获取用户uuid
	var user_uuid string
	if idstr, f := c.Get(conf.UserID); f {
		user_uuid = helper.Strval(idstr)
	}
	if user_uuid == "" {
		log.Error("CreateShareBatchHandler user uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取文件uuid user_file
	listJson, err := c.GetRawData()
	if err != nil {
		log.Error("CreateShareBatchHandler get json err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	taskList := &models.BatchTaskInfo{}
	err = json.Unmarshal([]byte(listJson), taskList)
	if err != nil {
		log.Error("CreateShareBatchHandler json parse err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取code
	code := c.PostForm(conf.Share_Code)
	if code == "" {
		log.Error("CreateShareBatchHandler code empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取过期时间并转化
	// 获取过期时间并转化
	expireStr := c.PostForm(conf.Share_Expire_Time)
	var expire sql.NullTime
	if expireStr != "" {
		tmpExpire, err := time.Parse("2006-01-02 15:04:05", expireStr)
		if err != nil {
			log.Error("CreateShareHandler expire_time invaild")
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.HTTP_INVALID_PARAMS_CODE,
				"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
			})
			return
		}
		expire = sql.NullTime{Time: tmpExpire, Valid: true}
	}
	// 生成share uuid
	share_uuid := helper.GenSid(user_uuid, code)
	// 创建分享文件夹将全部分享文件放入，文件夹所有者为管理员(uuid = 0)
	tmp := fmt.Sprintf("%s_%s", user_uuid, share_uuid)
	tmpUuid := helper.GenUserFid(conf.Administrator_Uuid, tmp)
	service.Mkdir(&models.UserFile{
		Uuid:      tmpUuid,
		User_Uuid: conf.Administrator_Uuid,
		Ext:       conf.Folder_Default_EXT,
		Name:      tmp,
	}, "")
	success := make([]string, 0)
	fail := make([]string, 0)
	for _, user_file_uuid := range taskList.Src {
		// 调用service层
		err = service.MoveObject(user_file_uuid, tmpUuid)
		if err != nil {
			log.Error("CreateShareBatchHandler create share record err ", err)
			fail = append(fail, user_file_uuid)
		} else {
			success = append(success, user_file_uuid)
		}
	}
	// 为每个file生成分享记录
	param := &models.CreateShareParams{
		Share_Uuid:     share_uuid,
		User_Uuid:      user_uuid,
		User_File_Uuid: tmpUuid,
		Code:           code,
		Expire:         expire,
	}
	// 调用service层
	err = service.SetShareLink(param)
	if err != nil {
		log.Error("CreateShareBatchHandler create share record err ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}

	// 成功
	log.Info("CreateShareBatchHandler success: ", share_uuid)
	c.JSON(http.StatusOK, gin.H{
		"code":     conf.SUCCESS_RESP_MESSAGE,
		"msg":      conf.SUCCESS_RESP_MESSAGE,
		"share_id": share_uuid,
		"location": tmpUuid, // 失败的文件直接copy到tmpuuid下即可
		"success":  success,
		"fail":     fail,
		"total":    len(taskList.Src),
	})
}
