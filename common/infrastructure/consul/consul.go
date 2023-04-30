package consul

import (
	"NetDesk/common/conf"
	"NetDesk/common/models"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/resolver"
)

type DiscoveryClientImpl struct {
	Registry *api.Client
}

func NewDiscoveryClientImpl(cfg *models.DiscoveryParam) (*DiscoveryClientImpl, error) {
	var res *DiscoveryClientImpl
	// 配置不生效使用默认配置
	if !cfg.Effect {
		r, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			return nil, err
		}
		res = &DiscoveryClientImpl{Registry: r}
	} else {
		// TODO 配置生效
	}
	// 注册resolver
	cb := &consulBuilder{}
	resolver.Register(cb)
	return res, nil
}

func (d *DiscoveryClientImpl) RegisterService(name, id, host string, port int) error {
	// Consul Client
	registry := d.Registry
	// 注册到 Consul，包含地址、端口信息，以及健康检查
	err := registry.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Port:    port,
		Address: host,
		Check: &api.AgentServiceCheck{
			TTL:     (conf.TTL + time.Second).String(),
			Timeout: time.Minute.String(),
		},
	})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("[%s] registry init err ", name))
	}
	return nil
}

func (d *DiscoveryClientImpl) KeepAlive(id string) {
	checkid := "service:" + id
	for range time.Tick(conf.TTL) {
		err := d.Registry.Agent().PassTTL(checkid, "")
		if err != nil {
			logrus.Error("[Noticeservice] PassTTL err: ", err)
		}
	}
}
