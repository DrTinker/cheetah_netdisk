package consul

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/resolver"
)

const (
	defaultPort = "8500"
)

var (
	errMissingAddr = errors.New("consul resolver: missing address")

	errAddrMisMatch = errors.New("consul resolver: invalied uri")

	errEndsWithColon = errors.New("consul resolver: missing port after port-separator colon")

	regexConsul, _ = regexp.Compile("^([A-z0-9.]+)(:[0-9]{1,5})?/([A-z_]+)$")
)

type consulBuilder struct{}

func (cb *consulBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	host, port, name, err := parseTarget(fmt.Sprintf("%s/%s", target.Authority, target.Endpoint()))
	if err != nil {
		logrus.Info("parse err")
		return nil, err
	}
	logrus.Info(fmt.Sprintf("consul service ==> host:%s, port%s, name:%s", host, port, name))
	cr := &consulResolver{
		address:              fmt.Sprintf("%s%s", host, port),
		name:                 name,
		cc:                   cc,
		disableServiceConfig: opts.DisableServiceConfig,
		Ch:                   make(chan int, 0),
	}
	go cr.watcher()
	return cr, nil

}

func (cb *consulBuilder) Scheme() string {
	return "consul"
}

func parseTarget(target string) (host, port, name string, err error) {

	if target == "" {
		return "", "", "", errMissingAddr
	}

	if !regexConsul.MatchString(target) {
		return "", "", "", errAddrMisMatch
	}

	groups := regexConsul.FindStringSubmatch(target)
	host = groups[1]
	port = groups[2]
	name = groups[3]
	if port == "" {
		port = defaultPort
	}
	return host, port, name, nil
}
