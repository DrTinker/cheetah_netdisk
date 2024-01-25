package main

import (
	"NetDisk/client"
	"NetDisk/start"
	"fmt"

	"github.com/gin-gonic/gin"
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
	start.InitMedia() // 媒体文件处理
	start.InitLOS()
	start.InitSocket()
}

func main() {
	// 初始化一个http服务对象
	r := gin.Default()

	start.RegisterRouter(r)

	// 开启http
	cfg, err := client.GetConfigClient().GetHttpConfig()
	if err != nil {
		panic(err)
	}

	r.Run(fmt.Sprintf("%s:%d", cfg.Address, cfg.Port))
}
