package models

type HttpConfig struct {
	Address string
	Port    int
}

type DBConfig struct {
	Type string
	User string
	Pwd  string
	IP   string
	Port int
	DB   string
}

type CacheConfig struct {
	IP   string
	Port int
	Pwd  string
}

type EmailConfig struct {
	Password string
	Name     string
	Email    string
	Address  string
	Port     int
}

type COSConfig struct {
	Domain    string
	Region    string
	SecretId  string
	SecretKey string
}

type LocalConfig struct {
	TmpPath  string
	FilePath string
}

// MQ连接配置
type MQConfig struct {
	Proto   string
	User    string
	Pwd     string
	Address string
	Port    int
}

// MinIO存储配置
type LOSConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}
