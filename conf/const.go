package conf

import "time"

// HTTP启动项
const App = "../../conf/app copy.ini"
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
const UserID = "userID"
const UserPWD = "userPwd"
const UserEmail = "userEmail"
const UserPhone = "userPhone"
const UserName = "userName"
const UserLevelNormal = 0 // 用户等级
const UserLevelVip = 1
const UserLevelSuper = 2
const UserNormalVolume = 10 * 1024 * 1024 * 1024 // 普通用户10GB空间
const UserVIPVolume = 20 * 1024 * 1024 * 1024    // 普通用户20GB空间

// 数据库
// 表名
const File_Pool_TB = "file_pool"
const User_TB = "user"
const Share_TB = "share"
const User_File_TB = "user_file"
const Trans_TB = "trans"
const Max_Conn = 100
const Max_Idle_Conn = 10
const Max_Idle_Time = time.Second * 30

// 数据表
// 表内行名称
const User_UUID_DB = "uuid"
const User_Email_DB = "email"
const User_Name_DB = "name"
const User_Now_Volume_DB = "now_volume"
const File_UUID_DB = "uuid"
const File_Hash_DB = "hash"
const File_Link_DB = "link"
const File_FileKey_DB = "file_key"
const File_Size_DB = "size"
const File_Thumbnail_DB = "thumbnail"
const File_Store_Type_DB = "store_type"
const User_File_ID_DB = "id"
const User_File_User_ID_DB = "user_uuid"
const User_File_UUID_DB = "uuid"
const User_File_Name_DB = "name"
const User_File_EXT_DB = "ext"
const User_File_Parent_DB = "parent_id"
const User_File_Pool_UUID_DB = "file_uuid"
const Share_UUID_DB = "uuid"
const Share_User_File_UUID_DB = "user_file_uuid"
const Share_User_Uuid_DB = "user_uuid"
const Share_Expire_DB = "expire_time"
const Trans_UUID_DB = "uuid"
const Trans_User_UUID_DB = "user_uuid"
const Trans_Status_DB = "status"
const Trans_IsDown_DB = "isdown"

// 邮件
const Email_Verify_MSG = "猎豹网盘 验证码"
const Forget_Password_MSG = "猎豹网盘 找回密码"
const Email_Verify_Name = "cheetah_net_desk"

// 验证码
const Code_Cache_Key = "verify_code_key"
const Code_Len = 6
const Code_Expire = 5 * time.Minute
const Code_Param_Key = "code"

// 传输状态
const Trans_Process = 0 // 上传中
const Trans_Success = 1 // 上传成功
const Trans_Fail = 2    // 上传失败（redis中key到期）
const Trans_Nil = -1

// 传输类型
const Upload_Mod = 0
const Download_Mod = 1

// 文件上传
const File_Part_Size_Max = 1024 * 1024 * 10 // 10MB
const Default_Thread_Pool_Size = 5
const Folder_Default_Size = 1
const Folder_Default_EXT = "folder"
const File_Exist_Flag = "exist"
const Publish_Retry_Times = 5

// 分块传输
const Trans_Part_Slice_Expire = time.Hour * 24 // 分块上传文件分块保存时间，用于断点续传

const Upload_Part_Info_Key = "Upload_Info" // 分块上传信息rediskey前缀
const Upload_Part_Info_ID_Key = "UploadID"
const Upload_Part_Info_CSize_Key = "ChunkSize"
const Upload_Part_Info_CCount_Key = "ChunkCount"
const Upload_Part_File_Info_Key = "FileInfo" // UploadObjectParams
const Upload_Part_Info_Fileds = 4            // redis 分块上传hash结构中前4条为配置信息，之后的kv才是已上传分块

const Download_Part_Info_Key = "Download_Info" // 分块下载信息rediskey前缀
const Download_Part_Info_ID_Key = "DownloadID"
const Download_Part_Info_CSize_Key = "ChunkSize"
const Download_Part_Info_CCount_Key = "ChunkCount"
const Download_Part_File_Path_Key = "FilePath"
const Download_Part_File_Size_Key = "Size"
const Download_Part_Ready_Key = "Ready" // 消息队列处理消息是否完成，
const Download_Part_Info_Fileds = 6     // redis 分块上传hash结构中前6条为配置信息，之后的kv才是已上传分块
const Download_Part_COS_Info_Fileds = 4

// 查询消息队列处理下载消息3中状态
const Download_Ready_Wait = "0"
const Download_Ready_Done = "1"
const Download_Ready_Abort = "2"

// mq相关
const Routing_Key = "cos"
const Exchange = "cheetah_netdesk"
const Transfer_COS_Queue = "cheetah_netdesk_trans_cos"
const Transfer_COS_Queue_Err = "cheetah_netdesk_trans_cos_err"
const Default_Content_Type = "text/plain"

// 文件存储
const Default_System_Prefix = "test" // 文件系统根目录，未来替换为root
const Default_Thumbnail_Prefix = "thumbnail"
const Default_System_parent = 0              // root对应ID
const Default_Sign_Expire = time.Minute * 15 // 签名默认有效时间
const Store_Type_COS = 0                     // 存储类型
const Store_Type_Tmp = 1                     // 临时存储
const Store_Type_Local = 2

// 通用配置
const Administrator_Uuid = "0"
const Default_Page_Size = 20

// 媒体处理
const Default_ThumbNail_Scale = "80x80"
const Default_ThumbNail_Frame = 24 // 视频缩略图截取帧，为第二秒的第一帧
const Default_Thumbnail_Ext = "jpg"

// websocket
const Buffer_Size = 1024
const Handshake_Timeout = 5 * time.Second

// 文件分享
const Share_All_Mod = 0    // 全部
const Share_Expire_Mod = 1 // 有效
const Share_Out_Mod = 2    // 过期
