package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NetDesk/client"
	"NetDesk/conf"
)

// 解析jwt，flag为true标识拦截不带token的请求
func JWT(flag bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int

		code = conf.HTTP_SUCCESS_CODE
		token := c.GetHeader(conf.JWTHeader)
		// 未携带token
		if token == "" {
			// 拦截
			if flag {
				c.Abort()
				return
			}
			// 标识未携带token登录
			c.Set(conf.JWTFlag, false)
			c.Next()
			return
		}
		// 解析token
		t, err := client.EncryptionClient.JwtDecode(token)
		if err != nil {
			log.Error("JWT error: ", err)
			code = conf.ERROR_AUTH_CHECK_TOKEN_FAIL_CODE
		} else if time.Now().Unix() > t.Expire {
			log.Error("JWT error: expire run out")
			code = conf.ERROR_AUTH_CHECK_TOKEN_TIMEOUT_CODE
		}
		// token无效
		if code != conf.HTTP_SUCCESS_CODE {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": code,
				"msg":  conf.JWT_ERROR_MESSAGE,
			})
			c.Abort()
			return
		}
		// 标识携带token登录
		c.Set(conf.User_ID, t.ID)
		c.Set(conf.User_Email, t.Email)
		c.Set(conf.User_PWD, t.Password)
		c.Set(conf.JWTFlag, true)
		c.Next()
	}
}
