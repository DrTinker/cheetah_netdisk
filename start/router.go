package start

import (
	"github.com/gin-gonic/gin"

	"NetDesk/handler/object"
	"NetDesk/handler/share"
	"NetDesk/handler/user"
	middleware "NetDesk/middleware"
)

// 初始化路由
func RegisterRouter(r *gin.Engine) {
	// 解决跨域
	r.Use(middleware.CORSHeader())
	u := r.Group("/user")
	{
		u.GET("/login", middleware.JWT(false), user.LoginHandler)
		u.POST("/register", user.RegisterHandler)
		u.GET("/email", user.EmailVerifyHandler)
	}

	o := r.Group("/object")
	{
		o.POST("/upload/total", middleware.JWT(true), middleware.FileCheck(), object.UploadHandler)
		o.POST("/upload_part/init", middleware.JWT(true), object.InitUploadPartHandler)
		o.POST("/upload_part/upload", middleware.JWT(true), object.UploadPartHandler)
		o.POST("/upload_part/complete", middleware.JWT(true), object.CompleteUploadPartHandler)
		o.POST("/mkdir", middleware.JWT(true), object.MakeDirHandler)
		o.GET("/list", middleware.JWT(true), object.GetFileListHandler)
		o.POST("/copy", middleware.JWT(true), object.CopyFileHandler)
		o.POST("/move", middleware.JWT(true), object.MoveFileHandler)
		o.GET("/info/path", middleware.JWT(true), object.GetFileInfoByPathHandler)
		o.POST("/rename", middleware.JWT(true), object.FileUpdateHandler)
		o.POST("/delete", middleware.JWT(true), object.FileDeleteHandler)
		o.GET("/token", middleware.JWT(true), object.GetTokenHandler)
	}

	// TODO 分享接口
	s := r.Group("/share")
	{
		s.POST("/create", middleware.JWT(true), share.CreateShareHandler)
		s.GET("/info", middleware.JWT(true), share.GetShareInfoHandler)
		s.POST("/copy", middleware.JWT(true), share.CopyFileByShareHandler)
		s.POST("/cancel", middleware.JWT(true), share.CancelShareHandler)
	}

	// 无效路由，处理自定义header
	r.NoRoute(middleware.CustomeHeader())
}
