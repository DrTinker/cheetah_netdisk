package middleware

import (
	"NetDesk/client"
	"NetDesk/conf"
	"NetDesk/helper"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// 通过md5值检测文件或文件夹是否存在，文件夹的md5值通过路径字符串生成
// mod为检查模式，0存在则直接拦截，1存在时进行标记并放行
func ExistCheck(mod int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取前端传入文件md5值
		md5 := c.PostForm(conf.File_Hash_Key)
		if md5 == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.ERROR_FILE_HASH_CODE,
				"msg":  conf.FILE_HASH_INVAILD_MESSAGE,
			})
			c.Abort()
			return
		}
		// gin不能重复读取body
		c.Set(conf.File_Hash_Key, md5)
		c.Set(conf.File_Quick_Upload_Key, false)
		// 通过数据库查询文件是否存在
		flag, uuid, err := client.GetDBClient().CheckFileExist(md5)
		if err != nil {
			log.Error("ExistCheck middleware file exist check err: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.ERROR_FILE_CHECK_CODE,
				"msg":  conf.FILE_CHECK_ERROR_MESSAGE,
			})
			c.Abort()
			return
		}
		if flag {
			switch mod {
			// mod 0 直接拦截
			case 0:
				log.Info("ExistCheck middleware file exist, hash: ", md5)
				c.JSON(http.StatusBadRequest, gin.H{
					"code": conf.ERROR_FILE_EXIST_CODE,
					"msg":  conf.FILE_EXIST_MESSAGE,
				})
				c.Abort()
				return
			// mod 1 放行并标记
			case 1:
				log.Info("ExistCheck middleware file exist, hash: ", md5)
				c.Set(conf.File_Uuid_Key, uuid)
				c.Set(conf.File_Quick_Upload_Key, true)
			}
		}
		c.Next()
	}
}

// 通过md5值检测客户端上传文件hash值是否合法
func FileCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// gin获取文件
		file, err := c.FormFile(conf.File_Form_Key)
		if err != nil {
			log.Error("UploadHandler err: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.HTTP_INVALID_PARAMS_CODE,
				"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
			})
			c.Abort()
			return
		}
		// 读取文件
		fd, err := file.Open()
		if err != nil {
			log.Error("UploadHandler file open err: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.HTTP_INVALID_PARAMS_CODE,
				"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
			})
			c.Abort()
			return
		}
		data, err := ioutil.ReadAll(fd)
		if err != nil {
			log.Error("UploadHandler file open err: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.HTTP_INVALID_PARAMS_CODE,
				"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
			})
			c.Abort()
			return
		}
		// 获取前端传入文件md5值
		md5 := c.PostForm(conf.File_Hash_Key)
		if md5 == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.ERROR_FILE_HASH_CODE,
				"msg":  conf.FILE_HASH_INVAILD_MESSAGE,
			})
			c.Abort()
			return
		}
		// gin不能重复读取body
		c.Set(conf.File_Hash_Key, md5)
		c.Set(conf.File_Form_Key, data)
		// 比较md5值
		hash := helper.CountMD5("", data, 1)
		if hash != md5 {
			log.Info("FileCheck middleware file md5 value invaild: ", md5)
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.ERROR_FILE_EXIST_CODE,
				"msg":  conf.FILE_EXIST_MESSAGE,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
