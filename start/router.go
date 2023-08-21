package start

import (
	"github.com/gin-gonic/gin"

	"NetDesk/handler/object"
	"NetDesk/handler/share"
	"NetDesk/handler/trans"
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
		u.GET("/profile", middleware.JWT(true), user.UserProfileHandler)
		u.GET("/info", middleware.JWT(true), user.UserInfoHandler)
		u.POST("/register", user.RegisterHandler)
		u.GET("/email", user.EmailVerifyHandler)
		u.GET("/forget", user.ForgetPwdHandler)
		u.POST("/rename", middleware.JWT(true), user.UpdateUserNameHandler)
	}

	t := r.Group("/trans")
	{
		t.GET("/info", middleware.JWT(true), trans.GetTransListHandler)

		t.POST("/del", middleware.JWT(true), trans.DelTransRecordHandler)
		t.POST("/del/batch", middleware.JWT(true), trans.DelTransBatchHandler)
		t.POST("/clear", middleware.JWT(true), trans.ClearTransListHandler)
		t.POST("/cancel/upload", middleware.JWT(true), trans.CancelUploadHandler)
		t.POST("/cancel/download", middleware.JWT(true), trans.CancelDownloadHandler)

		t.POST("/upload/total", middleware.JWT(true), middleware.FileCheck(), trans.UploadHandler)
		t.POST("/upload/init", middleware.JWT(true), trans.InitUploadHandler)
		t.POST("/upload/part", middleware.JWT(true), trans.UploadPartHandler)
		t.POST("/upload/complete", middleware.JWT(true), trans.CompleteUploadPartHandler)

		t.GET("/download/total", middleware.JWT(true), trans.DownloadFileHandler)
		t.GET("/download/init", middleware.JWT(true), trans.InitDownloadHandler)
		t.GET("/download/check", middleware.JWT(true), trans.CheckDownloadReadyHandler)
		t.GET("/download/part", middleware.JWT(true), trans.DownloadPartHandler)
		t.GET("/download/complete", middleware.JWT(true), trans.CompleteDownloadPartHandler)
	}

	o := r.Group("/object")
	{

		o.POST("/mkdir", middleware.JWT(true), object.MakeDirHandler)
		o.GET("/list", middleware.JWT(true), object.GetFileListHandler)
		o.POST("/copy", middleware.JWT(true), object.CopyFileHandler)
		o.POST("/move", middleware.JWT(true), object.MoveFileHandler)
		o.GET("/info", middleware.JWT(true), object.GetFileInfoHandler)
		o.POST("/rename", middleware.JWT(true), object.FileUpdateHandler)
		o.POST("/delete", middleware.JWT(true), object.FileDeleteHandler)

		o.GET("/token", middleware.JWT(true), object.GetTokenHandler)

		// batch
		o.POST("/batch/copy", middleware.JWT(true), object.CopyFileBatchHandler)
		o.POST("/batch/move", middleware.JWT(true), object.MoveFileBatchHandler)
		o.POST("/batch/delete", middleware.JWT(true), object.DeleteFileBatchHandler)
	}

	// TODO 分享接口
	s := r.Group("/share")
	{
		s.POST("/set", middleware.JWT(true), share.SetShareHandler)
		s.GET("/info", middleware.JWT(true), share.GetShareInfoHandler)
		s.GET("/list", middleware.JWT(true), share.GetShareListHandler)
		s.POST("/update", middleware.JWT(true), share.UpdateShareHandler)
		s.POST("/cancel", middleware.JWT(true), share.CancelShareBatchHandler)
	}

	// 无效路由，处理自定义header
	r.NoRoute(middleware.CustomeHeader())
}
