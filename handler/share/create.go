package share

import (
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"NetDisk/service"
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
	var userUuid string
	if idstr, f := c.Get(conf.UserID); f {
		userUuid = helper.Strval(idstr)
	}
	if userUuid == "" {
		log.Error("SetShareHandler user uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取UserFileUuid
	userFileUuid := c.PostForm(conf.ShareUserFileUuid)
	if userFileUuid == "" {
		log.Error("CreateShareHandler user file uuid empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取code
	code := c.PostForm(conf.ShareCode)
	if code == "" {
		log.Error("CreateShareHandler code empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取code
	fullName := c.PostForm(conf.ShareName)
	if fullName == "" {
		log.Error("CreateShareHandler name empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 获取过期时间并转化
	expireStr := c.PostForm(conf.ShareExpireTime)
	var expire sql.NullTime
	if expireStr != "" {
		tmpExpire, err := time.Parse("2006-01-02 15:04:05", expireStr)
		if err != nil {
			log.Error("CreateShareHandler ExpireTime invaild")
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.HTTP_INVALID_PARAMS_CODE,
				"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
			})
			return
		}
		expire = sql.NullTime{Time: tmpExpire, Valid: true}
	}

	var shareUuid string
	// 看是否传入
	shareUuid = c.PostForm(conf.ShareUuid)
	if shareUuid == "" {
		// 没传入则生成生成share uuid
		shareUuid = helper.GenSid(userUuid, code)
	}

	// 封装结构体
	param := &models.CreateShareParams{
		ShareUuid:    shareUuid,
		UserUuid:     userUuid,
		UserFileUuid: userFileUuid,
		Fullname:     fullName,
		Code:         code,
		Expire:       expire,
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
	log.Info("CreateShareHandler success: ", shareUuid)
	c.JSON(http.StatusOK, gin.H{
		"code":    conf.HTTP_SUCCESS_CODE,
		"msg":     conf.SUCCESS_RESP_MESSAGE,
		"shareID": shareUuid,
	})
}

func CreateShareBatchHandler(c *gin.Context) {
	// 获取用户uuid
	var UserUuid string
	if idstr, f := c.Get(conf.UserID); f {
		UserUuid = helper.Strval(idstr)
	}
	if UserUuid == "" {
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
	code := c.PostForm(conf.ShareCode)
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
	expireStr := c.PostForm(conf.ShareExpireTime)
	var expire sql.NullTime
	if expireStr != "" {
		tmpExpire, err := time.Parse("2006-01-02 15:04:05", expireStr)
		if err != nil {
			log.Error("CreateShareHandler ExpireTime invaild")
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.HTTP_INVALID_PARAMS_CODE,
				"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
			})
			return
		}
		expire = sql.NullTime{Time: tmpExpire, Valid: true}
	}
	// 生成share uuid
	ShareUuid := helper.GenSid(UserUuid, code)
	// 创建分享文件夹将全部分享文件放入，文件夹所有者为管理员(uuid = 0)
	tmp := fmt.Sprintf("%s_%s", UserUuid, ShareUuid)
	tmpUuid := helper.GenUserFid(conf.AdministratorUuid, tmp)
	service.Mkdir(&models.UserFile{
		Uuid:     tmpUuid,
		UserUuid: conf.AdministratorUuid,
		Ext:      conf.FolderDefaultExt,
		Name:     tmp,
	}, "")
	success := make([]string, 0)
	fail := make([]string, 0)
	for _, UserFileUuid := range taskList.Src {
		// 调用service层
		err = service.MoveObject(UserFileUuid, tmpUuid)
		if err != nil {
			log.Error("CreateShareBatchHandler create share record err ", err)
			fail = append(fail, UserFileUuid)
		} else {
			success = append(success, UserFileUuid)
		}
	}
	// 为每个file生成分享记录
	param := &models.CreateShareParams{
		ShareUuid:    ShareUuid,
		UserUuid:     UserUuid,
		UserFileUuid: tmpUuid,
		Code:         code,
		Expire:       expire,
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
	log.Info("CreateShareBatchHandler success: ", ShareUuid)
	c.JSON(http.StatusOK, gin.H{
		"code":     conf.SUCCESS_RESP_MESSAGE,
		"msg":      conf.SUCCESS_RESP_MESSAGE,
		"share_id": ShareUuid,
		"location": tmpUuid, // 失败的文件直接copy到tmpuuid下即可
		"success":  success,
		"fail":     fail,
		"total":    len(taskList.Src),
	})
}
