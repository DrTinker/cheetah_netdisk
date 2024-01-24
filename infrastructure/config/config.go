package config

import (
	"NetDisk/helper"
	"NetDisk/models"
	"errors"
	"fmt"

	"gopkg.in/ini.v1"
)

type ConfigClientImpl struct {
	Http   *models.HttpConfig
	DB     *models.DBConfig
	Email  *models.EmailConfig
	Cache  *models.CacheConfig
	COS    *models.COSConfig
	Local  *models.LocalConfig
	MQ     *models.MQConfig
	LOS    *models.LOSConfig
	source *ini.File
}

func NewConfigClientImpl() *ConfigClientImpl {
	return &ConfigClientImpl{
		Http:  &models.HttpConfig{},
		DB:    &models.DBConfig{},
		Email: &models.EmailConfig{},
		Cache: &models.CacheConfig{},
		COS:   &models.COSConfig{},
		Local: &models.LocalConfig{},
		MQ:    &models.MQConfig{},
		LOS:   &models.LOSConfig{},
	}
}

func (c *ConfigClientImpl) Load(path string) error {
	var err error
	//判断配置文件是否存在
	exists, _ := helper.PathExists(path)
	if !exists {
		return errors.New("config path not exists")
	}
	c.source, err = ini.Load(path)
	if err != nil {
		return err
	}
	return nil
}

func (c *ConfigClientImpl) GetHttpConfig() (*models.HttpConfig, error) {
	//判断配置是否加载成功
	if c.source == nil {
		return nil, errors.New("empty http config")
	}
	c.Http.Address = c.source.Section("HttpServer").Key("address").String()
	c.Http.Port = c.source.Section("HttpServer").Key("port").MustInt(8081)
	return c.Http, nil
}

func (c *ConfigClientImpl) GetDBConfig() (driver, source string, err error) {
	//判断配置是否加载成功
	if c.source == nil {
		return "", "", errors.New("empty db config")
	}
	c.DB.Type = c.source.Section("DB").Key("type").MustString("mysql")
	c.DB.User = c.source.Section("DB").Key("user").MustString("root")
	c.DB.IP = c.source.Section("DB").Key("ip").MustString("127.0.0.1")
	c.DB.Pwd = c.source.Section("DB").Key("pwd").MustString("xiaonajia123")
	c.DB.DB = c.source.Section("DB").Key("db").MustString("query_system")
	c.DB.Port = c.source.Section("DB").Key("port").MustInt(3306)

	driver = c.DB.Type
	source = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True",
		c.DB.User, c.DB.Pwd, c.DB.IP, c.DB.Port, c.DB.DB)

	return driver, source, nil
}

func (c *ConfigClientImpl) GetEmailConfig() (*models.EmailConfig, error) {
	//判断配置是否加载成功
	if c.source == nil {
		return nil, errors.New("empty email config")
	}

	c.Email.Address = c.source.Section("Email").Key("address").MustString("smtp.163.com")
	c.Email.Password = c.source.Section("Email").Key("pwd").MustString("ZFXJQMJPMAGACPKU")
	c.Email.Email = c.source.Section("Email").Key("email").MustString("cheetah_net_disk@163.com")
	c.Email.Name = c.source.Section("Email").Key("name").MustString("cheetah_net_disk")
	c.Email.Port = c.source.Section("Email").Key("port").MustInt(587)

	return c.Email, nil
}

func (c *ConfigClientImpl) GetCacheConfig() (addr, pwd string, err error) {
	//判断配置是否加载成功
	if c.source == nil {
		return "", "", errors.New("empty cache config")
	}
	section := c.source.Section("Cache")
	c.Cache = &models.CacheConfig{}
	c.Cache.IP = section.Key("address").MustString("127.0.0.1")
	c.Cache.Port = section.Key("port").MustInt(6379)
	c.Cache.Pwd = section.Key("pwd").String()

	addr = fmt.Sprintf("%s:%d", c.Cache.IP, c.Cache.Port)
	pwd = c.Cache.Pwd
	return addr, pwd, nil
}

func (c *ConfigClientImpl) GetCOSConfig() (*models.COSConfig, error) {
	//判断配置是否加载成功
	if c.source == nil {
		return nil, errors.New("empty cos config")
	}
	section := c.source.Section("COS")
	c.COS.SecretId = section.Key("SecretId").String()
	c.COS.SecretKey = section.Key("SecretKey").String()
	c.COS.Domain = section.Key("domain").String()
	c.COS.Region = section.Key("region").String()

	return c.COS, nil
}

func (c *ConfigClientImpl) GetLocalConfig() (*models.LocalConfig, error) {
	//判断配置是否加载成功
	if c.source == nil {
		return nil, errors.New("empty cos config")
	}
	section := c.source.Section("Local")
	c.Local.TmpPath = section.Key("tmppath").String()
	c.Local.FilePath = section.Key("filepath").String()

	return c.Local, nil
}

func (c *ConfigClientImpl) GetMQConfig() (*models.MQConfig, error) {
	//判断配置是否加载成功
	if c.source == nil {
		return nil, errors.New("empty mq config")
	}

	c.MQ.Address = c.source.Section("MQ").Key("address").MustString("127.0.0.1")
	c.MQ.Pwd = c.source.Section("MQ").Key("pwd").MustString("guest")
	c.MQ.Proto = c.source.Section("MQ").Key("proto").MustString("amqp")
	c.MQ.User = c.source.Section("MQ").Key("user").MustString("guest")
	c.MQ.Port = c.source.Section("MQ").Key("port").MustInt(5672)

	return c.MQ, nil
}

func (c *ConfigClientImpl) GetLOSConfig() (*models.LOSConfig, error) {
	//判断配置是否加载成功
	if c.source == nil {
		return nil, errors.New("empty los config")
	}
	section := c.source.Section("LOS")
	c.LOS.Endpoint = section.Key("endpoint").String()
	c.LOS.AccessKeyID = section.Key("accessKeyID").String()
	c.LOS.SecretAccessKey = section.Key("secretAccessKey").String()
	c.LOS.UseSSL = section.Key("useSSL").MustBool()

	return c.LOS, nil
}
