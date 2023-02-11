package middleware

import (
	"NetDisk/conf"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 解决跨域问题
func CORSHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		// 响应类型
		c.Header("Access-Control-Allow-Headers", "content-type,token,id,Authorization")
		// 响应头设置
		c.Header("Access-Control-Request-Headers", "Origin, X-Requested-With, content-Type, Accept, Authorization")

		c.Next()
	}
}

func CustomeHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": conf.HTTP_SUCCESS_CODE,
			"msg":  conf.SUCCESS_RESP_MESSAGE,
		})
	}
}
