package main

import (
	"NetDisk/handler/trans"
	"NetDisk/start"
	"fmt"
)

// 初始化
func init() {
	start.InitConfig() // 加载配置
	start.InitDB()     // 数据库
	start.InitCache()  // 缓存
	start.InitCOS()    //对象存储
	start.InitLOS()
	start.InitMQ()
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Runtime panic caught: %v\n", err)
			trans.TransferObjectHandler()
		}
	}()
	fmt.Printf("running transfer service\n")

	trans.TransferObjectHandler()
}
