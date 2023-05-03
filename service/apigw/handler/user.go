package handler

import (
	"NetDesk/common/conf"
	"NetDesk/common/models"
	"NetDesk/service/apigw/logic"
	"NetDesk/service/apigw/proto/user"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func SignInHandler(c *gin.Context) {
	// 初始化user struct
	u := models.Login{}
	var token string
	// 处理jwt token
	if c.GetBool(conf.JWTFlag) {
		if email, ok := c.Get(conf.User_Email); ok && email != nil {
			u.Email, _ = email.(string)
		}
		if pwd, ok := c.Get(conf.User_PWD); ok && pwd != nil {
			u.Password, _ = pwd.(string)
		}
	} else {
		err := c.ShouldBind(&u)
		if err != nil {
			log.Error("SignInHandler err: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"code": conf.HTTP_INVALID_PARAMS_CODE,
				"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
			})
		}
	}
	email := u.Email
	pwd := u.Password

	l, err := logic.NewUserLogic()
	if err != nil {
		log.Error("SignInHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}

	resp, token, err := l.UserSignIn(&user.UserSignInReq{
		Email:    email,
		Password: pwd,
	}, c.GetBool(conf.JWTFlag))
	if err != nil {
		log.Error("SignInHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}

	// 返回成功
	log.Info("SignInHandler success: %v", u.User_UUID)
	// 返回值去掉密码字段
	resp.UserInfo.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"code":  resp.Resp.Code,
		"msg":   resp.Resp.RespMsg,
		"data":  resp.UserInfo,
		"token": token,
	})
}

func SignUpHandler(c *gin.Context) {
	// 初始化user struct
	user := &models.User{}
	err := c.ShouldBind(&user)
	if err != nil {
		log.Error("SignUpHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.HTTP_INVALID_PARAMS_CODE,
			"msg":  conf.HTTP_INVALID_PARAMS_MESSAGE,
		})
		return
	}
	// 调用user服务
	l, err := logic.NewUserLogic()
	if err != nil {
		log.Error("SignUpHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}

	// 获取code
	code := c.Query(conf.Code_Param_Key)
	resp, err := l.UserSignUp(user, code)
	if err != nil {
		log.Error("SignUpHandler err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code": conf.SERVER_ERROR_CODE,
			"msg":  conf.SERVER_ERROR_MSG,
		})
		return
	}

	// 返回成功
	log.Info("SignUpHandler success: ", resp.UserUuid)
	c.JSON(http.StatusOK, gin.H{
		"code":    resp.Resp.Code,
		"msg":     resp.Resp.RespMsg,
		"user_id": resp.UserUuid,
	})
}
