package user

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/models"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func LoginHandler(c *gin.Context) {
	// 初始化user struct
	u := models.Login{}
	var token string
	// 处理jwt token
	if c.GetBool(conf.JWTFlag) {
		if email, ok := c.Get(conf.UserEmail); ok && email != nil {
			u.Email, _ = email.(string)
		}
		if pwd, ok := c.Get(conf.UserPWD); ok && pwd != nil {
			u.Password, _ = pwd.(string)
		}
	} else {
		err := c.ShouldBind(&u)
		if err != nil {
			log.Error("LoginHandler err: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.HTTP_INVALID_PARAMS_CODE,
				"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
			})
		}
	}
	// email不能为空
	if u.Email == "" || u.Password == "" {
		log.Error("LoginHandler err: empty email or pwd")
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	email := u.Email
	pwd := u.Password

	info, err := client.GetDBClient().GetUserByEmail(email)
	if err != nil && err != conf.DBNotFoundError {
		log.Error("LoginHandler pwd err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}
	if err == conf.DBNotFoundError || info.Password != pwd {
		log.Error("LoginHandler pwd err: ", err)
		c.JSON(http.StatusOK, gin.H{
			"code": conf.ERROR_LOGIN_CODE,
			"msg":  conf.LOGIN_ERROR_MESSAGE,
		})
		return
	}

	// 未携带jwt则为初次登录
	if !c.GetBool(conf.JWTFlag) {
		token, _ = client.EncryptionClient.JwtEncode(models.Token{
			ID:       info.Uuid,
			Email:    email,
			Password: pwd,
			Expire:   0,
		})
	}

	// 返回成功
	log.Info("LoginHandler success: ", u.UserUUID)
	// 返回值去掉密码字段
	info.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"code":  conf.HTTP_SUCCESS_CODE,
		"msg":   conf.SUCCESS_RESP_MESSAGE,
		"data":  info,
		"token": token,
	})
}
