package start

import (
	"github.com/gin-gonic/gin"

	"NetDisk/handler/object"
	"NetDisk/handler/user"
	middleware "NetDisk/middleware"
)

// 初始化路由
func RegisterRouter(r *gin.Engine) {
	u := r.Group("/user")
	{
		u.GET("/login", middleware.JWT(false), user.LoginHandler)
		u.POST("/register", user.RegisterHandler)
		u.GET("/email", user.EmailVerifyHandler)
	}

	o := r.Group("/object")
	{
		o.POST("/upload", middleware.JWT(true), middleware.ExistCheck(1), object.UploadHandler)
		o.POST("/mkdir", middleware.JWT(true), middleware.ExistCheck(0), object.MakeDirHandler)
		o.GET("/list", middleware.JWT(true), object.GetFileListHandler)
	}
}
