package main

import (
	"NetDesk/handler/object"
	"NetDesk/start"
	"fmt"
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
}

func main() {
	fmt.Printf("running transfer service")
	object.TransferObjectHandler()
}
