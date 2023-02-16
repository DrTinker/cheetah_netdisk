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
		o.POST("/copy", middleware.JWT(true), object.CopyFileHandler)
		o.POST("/move", middleware.JWT(true), object.MoveFileHandler)
		o.GET("/info/path", middleware.JWT(true), object.GetFileInfoByPathHandler)
		// TODO 文件改名、删除
		o.POST("/rename", middleware.JWT(true), object.FileUpdateHandler)
		o.POST("/delete", middleware.JWT(true), object.FileDeleteHandler)
	}
}
