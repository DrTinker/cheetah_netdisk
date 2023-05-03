package main

import (
	"NetDesk/common/client"
	"NetDesk/common/helper"
	"NetDesk/common/start"
	"NetDesk/service/notice/handler"
	"NetDesk/service/notice/proto/notice"
	"flag"
	"net"
	"strconv"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// 初始化
func init() {
	start.InitConfig() // 加载配置
	start.InitDiscoveryClient()
	start.InitCache()
	start.InitMsg()
}

func main() {
	// 读取命令行参数
	host := flag.String("h", "127.0.0.1", "host")
	port := flag.Int("p", 50052, "port")
	flag.Parse()

	lis, err := net.Listen("tcp", net.JoinHostPort(*host, strconv.Itoa(*port)))
	if err != nil {
		logrus.Error("[Noticeservice] failed to listen: ", err)
	}

	id := helper.GenServiceID("Noticeservice", *port)
	logrus.Info("[Noticeservice] service id: ", id)

	// 注册到 Consul，包含地址、端口信息，以及健康检查
	err = client.GetDiscoveryClient().RegisterService("Noticeservice", id, *host, *port)
	if err != nil {
		logrus.Error("[Noticeservice] ServiceRegister err: ", err)
	}
	// keepalive
	go func() {
		client.GetDiscoveryClient().KeepAlive(id)
	}()

	// 服务注册
	s := grpc.NewServer()
	notice.RegisterNoticeserviceServer(s, handler.NoticeService)

	if err := s.Serve(lis); err != nil {
		logrus.Error("[Noticeservice] fail to serve: ", err)
	}
}
