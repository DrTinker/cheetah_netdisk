package main

import (
	"NetDesk/common/client"
	"NetDesk/common/helper"
	"NetDesk/common/start"
	"NetDesk/service/user/handler"
	"NetDesk/service/user/proto/user"
	"flag"
	"net"
	"strconv"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// 初始化
func init() {
	start.InitConfig() // 加载配置
	start.InitDB()     // 数据库
	start.InitDiscoveryClient()
}

func main() {
	// 读取命令行参数
	host := flag.String("h", "127.0.0.1", "host")
	port := flag.Int("p", 50051, "port")
	flag.Parse()

	lis, err := net.Listen("tcp", net.JoinHostPort(*host, strconv.Itoa(*port)))
	if err != nil {
		logrus.Error("[Userservice] failed to listen: ", err)
	}

	id := helper.GenServiceID("Userservice", port)
	logrus.Info("[Userservice] service id: ", id)

	// 注册到 Consul，包含地址、端口信息，以及健康检查
	err = client.GetDiscoveryClient().RegisterService("Userservice", id, *host, *port)
	if err != nil {
		logrus.Error("[Userservice] ServiceRegister err: ", err)
	}
	// keepalive
	go func() {
		client.GetDiscoveryClient().KeepAlive(id)
	}()

	// 服务注册
	s := grpc.NewServer()
	user.RegisterUserServiceServer(s, handler.UserService)

	if err := s.Serve(lis); err != nil {
		logrus.Error("[Userservice] fail to serve: ", err)
	}
}
