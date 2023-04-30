package consul

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
)

// 实现resolver接口
type consulResolver struct {
	address              string
	wg                   sync.WaitGroup
	cc                   resolver.ClientConn
	name                 string
	disableServiceConfig bool
	Ch                   chan int
}

func (cr *consulResolver) ResolveNow(opt resolver.ResolveNowOptions) {
	cr.Ch <- 1
}

func (cr *consulResolver) Close() {
}

func (cr *consulResolver) watcher() {
	fmt.Printf("calling [%s] consul watcher\n", cr.name)
	config := api.DefaultConfig()
	config.Address = cr.address
	client, err := api.NewClient(config)
	if err != nil {
		fmt.Printf("error create consul client: %v\n", err)
		return
	}
	t := time.NewTicker(2000 * time.Millisecond)
	defer func() {
		fmt.Println("defer done")
	}()
	for {
		select {
		case <-t.C:
			//fmt.Println("定时")
		case <-cr.Ch:
			//fmt.Println("ch call")
		}
		//api添加了 lastIndex   consul api中并不兼容附带lastIndex的查询
		services, _, err := client.Health().Service(cr.name, "", true, &api.QueryOptions{})
		if err != nil {
			fmt.Printf("error retrieving instances from Consul: %v", err)
		}

		newAddrs := make([]resolver.Address, 0)
		for _, service := range services {
			addr := net.JoinHostPort(service.Service.Address, strconv.Itoa(service.Service.Port))
			newAddrs = append(newAddrs, resolver.Address{
				Addr: addr,
				//type：不能是grpclb，grpclb在处理链接时会删除最后一个链接地址，不用设置即可 详见=> balancer_conn_wrappers => updateClientConnState
				ServerName: service.Service.Service,
			})
		}
		//cr.cc.NewAddress(newAddrs)
		//cr.cc.NewServiceConfig(cr.name)
		cr.cc.UpdateState(resolver.State{Addresses: newAddrs})
	}
}
