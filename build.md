# 服务端部署

暂时采用docker容器的方式部署，minio的部署有很多坑



## 1. 构建镜像

**在顶层目录执行**

```shell
# web服务
docker build -f deploy/service_dc/main-dockerfile -t main-image .
# 文件转移服务(从私有云转移到公有云)
docker build -f deploy/service_dc/transfer-dockerfile -t transfer-image .
```



## 2. 创建容器

#### 部署mysql + redis + rabbitmq

```shell
# 运行redis容器 redis-server --appendonly yes表示开启持久化
docker run --name myredis -p 6379:6379 -d redis redis-server --appendonly yes

# 运行mysql容器
docker run --name mymysql -e MYSQL_ROOT_PASSWORD=pwd -d -p 3306:3306 mysql

# 运行rabbitmq:management(带客户端的版本) user和pass不要用guest，否则只有本机能使用
docker run \
    --name myrabbit --hostname my-rabbit \
    -p 15672:15672 -p 5672:5672 \
    -e RABBITMQ_DEFAULT_USER=root -e RABBITMQ_DEFAULT_PASS=pwd \
    -d rabbitmq:management
```

#### 导入数据库表

导入netdisk.sql

#### 创建队列和交换机

登录rabbitmq web管理端 ip:15672
创建queue: cheetah_NetDisk_trans_cos
创建exhange: cheetah_NetDisk type: direct
建立binding: routing key: cos

#### 部署minio(https)

minio默认只支持http请求，开启https有两种选择

1. 为minio配置ssl证书（本文选择此方案）
2. nginx反向代理

生成自签名证书(有域名及ssl证书忽略此步)

```shell
# 下载证书生成工具 这里推荐certgen https://github.com/minio/certgen/releases/tag/v0.0.2

# 执行命令 生成私钥与证书: private.key public.crt
certgen -ca -host "192.168.0.1,172.17.0.3" # (服务器内网ip,docker容器虚拟ip)

# docker容器虚拟ip可通过 (docker inspect 容器id) 命令查询
```
注意自签名证书只能用于容器间的https传输，客户端依然无法通过https访问，因此如果是自签名可以不必开启https(后文中复制证书的步骤也可忽略)

创建容器

```shell
# ~/minio/data替换为宿主机存储minio数据的路径
# ~/minio/config替换为宿主机保存私钥与证书的路径
docker run \
   -p 9000:9000 \
   -p 9001:9001 \
   --name myminio \
   -v ~/minio/data:/data \
   -v ~/minio/config:/root/.minio \
   -e "MINIO_ROOT_USER=admin" \
   -e "MINIO_ROOT_PASSWORD=pwd" \
   -d quay.io/minio/minio server /data --console-address ":9001"
```

测试minio

在浏览器输入https://ip:9001访问minio客户端

#### 部署后端服务

部署main 服务端口8081

```shell
# ~/cfg为宿主机配置文件存储路径，配置文件须命名为 app.ini
docker run \
    -p 8081:8081 --name main-server \
    -v ~/cfg:/build/cfg \
    --link mymysql:emysql --link myredis:eredis --link myrabbit:erabbit --link myminio:eminio \
    -d main-image

# 把生成的证书复制到容器中，如果证书不是自签名的则忽略
# /root/minio/config/certs/ 须改成宿主机中存放证书的路径
docker cp /root/minio/config/certs/private.key main-server:/etc/ssl/certs/
docker cp /root/minio/config/certs/public.crt main-server:/etc/ssl/certs/
```

部署transfer 不暴露端口

```shell
# ~/cfg为宿主机配置文件存储路径, 配置文件须命名为 app.ini
docker run \
    --name transfer-server \
    -v ~/cfg:/build/cfg \
    --link mymysql:emysql --link myredis:eredis --link myrabbit:erabbit --link myminio:eminio \
    -d transfer-image

# 把生成的证书复制到容器中，如果证书不是自签名的则忽略
# /root/minio/config/certs/ 须改成宿主机中存放证书的路径
docker cp /root/minio/config/certs/private.key transfer-server:/etc/ssl/certs/
docker cp /root/minio/config/certs/public.crt transfer-server:/etc/ssl/certs/
```

配置文件 app.ini

```ini
[ApiService]
address = ip
port = 8081
[DB]
type = mysql
user = root
pwd = pwd
ip = emysql
port = 3306
db = netdisk
[Email]
pwd = pwd
name = name
email = youremail@xxx.com
address = smtp.xxx.com
port = 587
[Cache]
address = eredis
port = 6379
pwd = pwd
[COS]
domain = https://my_domain
region = https://my_region
SecretId = mySecretId
SecretKey = mySecretKey
[Local] # 暂存文件的服务端本地磁盘路径
tmppath = mytmppath
filepath = myfilepath
[MQ]
proto = amqp
user = root
pwd = pwd
address = erabbit
port = 5672
[LOS]
endpoint = eminio:9000
accessKeyID = myaccessKeyID
secretAccessKey = mysecretAccessKey
# minio 是否支持 https 访问, 没有域名的建议为 false
useSSL = false
```



## 3. 测试

注意关闭服务器防火墙或是开放对应端口

发送http请求：http://ip:8081/hello 收到如下请求

```json
{"code":10000,"msg":"No router!"}
```

