package main

import (
	"NetDesk/common/client"
	"NetDesk/common/helper"
	"NetDesk/common/start"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"NetDesk/handler/object"
	"NetDesk/handler/share"
	"NetDesk/handler/user"
	middleware "NetDesk/middleware"
	"NetDesk/service/apigw/handler"
)

// 初始化
func init() {
	start.InitConfig() // 加载配置
	start.InitDB()     // 数据库
	start.InitCache()  // 缓存
	start.InitJWT()
	start.InitMsg() // 邮件系统
	start.InitCOS() //对象存储
	start.InitMQ()
	start.InitDiscoveryClient()
}

// 初始化路由
func RegisterRouter(r *gin.Engine) {
	// 解决跨域
	r.Use(middleware.CORSHeader())
	u := r.Group("/user")
	{
		u.GET("/login", middleware.JWT(false), handler.SignInHandler)
		u.POST("/register", handler.SignUpHandler)
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
		o.GET("/download/total", middleware.JWT(true), object.DownloadFileHandler)
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

func main() {
	// 初始化一个http服务对象
	r := gin.Default()

	RegisterRouter(r)

	cfg, err := client.GetConfigClient().GetHttpConfig()
	if err != nil {
		panic(err)
	}
	id := helper.GenServiceID("ApiGW", cfg.Port)
	logrus.Info("[ApiGW] service id: ", id)

	// 注册到 Consul，包含地址、端口信息，以及健康检查
	err = client.GetDiscoveryClient().RegisterService("ApiGW", id, cfg.Address, cfg.Port)
	if err != nil {
		logrus.Error("[ApiGW] ServiceRegister err: ", err)
	}
	// keepalive
	go func() {
		client.GetDiscoveryClient().KeepAlive(id)
	}()
	// run会阻塞，应在run前注册consul
	r.Run(fmt.Sprintf("%s:%d", cfg.Address, cfg.Port))
}
