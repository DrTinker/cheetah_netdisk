# 猎豹网盘 Cheetah Netdisk

*based on go1.21 (go1.18 at least)*



## Introduction

猎豹网盘是个人开发的一款开源的网盘系统，采用前后端分离架构

此仓库为后端代码，采用Gin作为开发框架，目前主要实现

1. 用户登录注册：邮箱验证，JWT鉴权
2. 用户文件系统：支持目录及文件的创建、删除、复制、移动等操作
3. 文件上传下载：大文件分片传输，断点续传，秒传
4. 媒体文件处理：生成图片、视频文件缩略图
5. 云端文件分享：生成已上传文件分享链接，设置提取码及过期时间，通过口令提取分享文件
6. 混合云存储：采用公有云(腾讯云COS)和私有云(minio)进行文件存储，不同存储间异步转移

前端采用flutter框架开发，仓库地址：https://github.com/DrTinker/NetDisk_front



## Code

```
|-- client			# 第三方依赖接口定义(db,cache,mq等)
|-- conf			# 常量、配置等
|-- deploy			# 部署相关文件 shell dockerfile等
|-- handler			# handler层，处理http请求，调用service层
|-- helper			# 各类与业务无关的方法
|-- infrastructure  # 第三方依赖的具体实现，实现client中定义的接口
|-- middleware		# gin中间件
|-- models			# 各类struct
|-- run				# 各服务运行入口
|   |-- server
|   `-- transfer
|-- service			# 实现业务逻辑，调用client层
`-- start			# 项目初始化相关逻辑
```



## Install

详见build.md



## Usage

接口文档整理中...

可参考 start/router.go中的定义



## TODO

- [x] 实现基本功能(用户信息、文件系统、简单上传下载、云端文件分享)
- [x] 上传下载优化(分片、秒传、断点续传)
- [x] 混合云存储
- [x] 媒体文件处理(视频压缩转码待实现)
- [ ] 接入Elasticsearch实现文件检索
- [ ] 微服务改造，集群部署
- [ ] 规范日志输出，增加日志收集
- [ ] 文档整理，测试与持续迭代