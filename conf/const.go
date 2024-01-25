package conf

import "time"

// HTTP启动项
const App = "./cfg/app.ini"
const HttpServer = "HttpServer"

// TTL
const MQTTL = 2 * time.Second

// jwt
const JWTKeyValue = "cheetah_net_disk"
const JWTExpireValue = 3600000
const JWTFlag = "jwt_flag"        // 标识本次登录是否携带JWT
const JWTClaims = "jwt_claims"    // 存储jwt的声明字段
const JWTHeader = "Authorization" // jwt请求头标识字段

// user
const UserLevelNormal = 0 // 用户等级
const UserLevelVip = 1
const UserLevelSuper = 2
const UserNormalVolume = 10 * 1024 * 1024 * 1024 // 普通用户10GB空间
const UserVIPVolume = 20 * 1024 * 1024 * 1024    // 普通用户20GB空间

// 数据库
// 表名
const FilePoolTB = "file_pool"
const UserTB = "user"
const ShareTB = "share"
const UserFileTB = "user_file"
const TransTB = "trans"
const MaxConn = 100
const MaxIdleConn = 10
const MaxIdleTime = time.Second * 30

// 数据表
// 表内行名称
const UserUuidDB = "uuid"
const UserEmailDB = "email"
const UserNameDB = "name"
const UserNowVolumeDB = "now_volume"
const FileUuidDB = "uuid"
const FileHashDB = "hash"
const FileLinkDB = "link"
const FileFileKeyDB = "file_key"
const FileSizeDB = "size"
const FileThumbnailDB = "thumbnail"
const FileStoreTypeDB = "store_type"
const UserFileIdDB = "id"
const UserFileUserIdDB = "user_uuid"
const UserFileUuidDB = "uuid"
const UserFileNameDB = "name"
const UserFileExtDB = "ext"
const UserFileParentDB = "parent_id"
const UserFilePoolUuidDB = "file_uuid"
const ShareUuidDB = "uuid"
const ShareUserFileUuidDB = "user_file_uuid"
const ShareUserUuidDB = "user_uuid"
const ShareExpireDB = "expire_time"
const TransUuidDB = "uuid"
const TransUserUuidDB = "user_uuid"
const TransStatusDB = "status"
const TransIsDownDB = "isdown"

// 邮件
const EmailVerifyMsg = "猎豹网盘 验证码"
const ForgetPasswordMsg = "猎豹网盘 找回密码"
const EmailVerifyName = "cheetah_net_disk"

// 验证码
const CodeCacheKey = "verify_code_key"
const CodeLen = 6
const CodeExpire = 5 * time.Minute
const CodeParamKey = "code"

// 传输状态
const TransProcess = 0 // 上传中
const TransSuccess = 1 // 上传成功
const TransFail = 2    // 上传失败（redis中key到期）
const TransNil = -1

// 传输类型
const UploadMod = 0
const DownloadMod = 1

// 文件上传
const FilePartSizeMax = 1024 * 1024 * 10 // 10MB
const DefaultThreadPoolSize = 5
const FolderDefaultSize = 1
const FolderDefaultExt = "folder"
const FileExistFlag = "exist"
const PublishRetryTimes = 5

// 分块传输
const Trans_Part_Slice_Expire = time.Hour * 24 // 分块上传文件分块保存时间，用于断点续传

const UploadPartInfoKey = "Upload_Info" // 分块上传信息rediskey前缀
const UploadPartInfoIDKey = "UploadID"
const UploadPartInfoCSizeKey = "ChunkSize"
const UploadPartInfoCCountKey = "ChunkCount"
const UploadPartFileInfoKey = "FileInfo" // UploadObjectParams
const UploadPartInfoFileds = 4           // redis 分块上传hash结构中前4条为配置信息，之后的kv才是已上传分块

const DownloadPartInfoKey = "Download_Info" // 分块下载信息rediskey前缀
const DownloadPartInfoIDKey = "DownloadID"
const DownloadPartInfoCSizeKey = "ChunkSize"
const DownloadPartInfoCCountKey = "ChunkCount"
const DownloadPartFilePathKey = "FilePath"
const DownloadPartFileSizeKey = "Size"
const DownloadPartReadyKey = "Ready" // 消息队列处理消息是否完成，
const DownloadPartInfoFileds = 6     // redis 分块上传hash结构中前6条为配置信息，之后的kv才是已上传分块
const DownloadPartCOSInfoFileds = 4

// 查询消息队列处理下载消息3中状态
const DownloadReadyWait = "0"
const DownloadReadyDone = "1"
const DownloadReadyAbort = "2"

// mq相关
const RoutingKey = "cos"
const Exchange = "cheetah_NetDisk"
const TransferCOSQueue = "cheetah_NetDisk_trans_cos"
const TransferCOSQueueErr = "cheetah_NetDisk_trans_cos_err"
const DefaultContentType = "text/plain"

// 文件存储
const FilePrefix = "test" // 文件系统根目录，未来替换为root
const ThumbnailPrefix = "thumbnail"
const TmpPrefix = "tmp"
const DefaultLocalPrefix = "./Local storage"
const DefaultLOSBucket = "netdisk"
const DefaultLOSExpire = 1                 // day 默认私有云保存1天
const DefaultSystemparent = 0              // root对应ID
const DefaultSignExpire = time.Minute * 15 // 签名默认有效时间
const StoreTypeCOS = 0                     // COS持久存储
const StoreTypeLOS = 1                     // LOS临时存储

// 通用配置
const AdministratorUuid = "0"
const DefaultPageSize = 20

// 媒体处理
const DefaultThumbnailScale = "80x80"
const DefaultThumbnailFrame = 24 // 视频缩略图截取帧，为第二秒的第一帧
const DefaultThumbnailExt = "jpg"

// websocket
const BufferSize = 1024
const HandshakeTimeout = 5 * time.Second

// 文件分享
const ShareAllMod = 0    // 全部
const ShareExpireMod = 1 // 有效
const ShareOutMod = 2    // 过期
