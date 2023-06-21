package conf

import "time"

// HTTP启动项
const App = "D:\\Program\\Projects\\netdesk_system\\merge\\conf\\app.ini"
const HttpServer = "HttpServer"

// jwt
const JWTKeyValue = "cheetah_net_disk"
const JWTExpireValue = 3600000
const JWTFlag = "jwt_flag"        // 标识本次登录是否携带JWT
const JWTClaims = "jwt_claims"    // 存储jwt的声明字段
const JWTHeader = "Authorization" // jwt请求头标识字段

// user
const User_ID = "user_id"
const User_PWD = "user_pwd"
const User_Email = "user_email"
const User_Level_normal = 0 // 用户等级
const User_Level_vip = 1
const User_Level_super = 2
const User_Normal_Volume = 10 * 1024 * 1024 * 1024 // 普通用户10GB空间
const User_VIP_Volume = 10 * 1024 * 1024 * 1024    // 普通用户20GB空间

// 数据库
// 表名
const File_Pool_TB = "file_pool"
const User_TB = "user"
const Share_TB = "share"
const User_File_TB = "user_file"
const Max_Conn = 100
const Max_Idle_Conn = 10
const Max_Idle_Time = time.Second * 30

// 数据表
// 表内行名称
const User_UUID_DB = "uuid"
const User_Email_DB = "email"
const User_Now_Volume_DB = "now_volume"
const File_UUID_DB = "uuid"
const File_Hash_DB = "hash"
const File_Link_DB = "link"
const File_Path_DB = "path"
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
const Share_Click_Num_DB = "click_num"

// 邮件
const Email_Verify_MSG = "Cheetah Net Desk验证码"
const Email_Verify_Name = "cheetah_net_desk"

// 验证码
const Code_Cache_Key = "verify_code_key"
const Code_Len = 6
const Code_Expire = 5 * time.Minute
const Code_Param_Key = "code"

// 文件上传
const File_Part_Size_Max = 1024 * 1024 // 1MB
const Default_Thread_Pool_Size = 5
const Folder_Default_Size = 1
const Folder_Default_EXT = "folder"
const File_Exist_Flag = "exist"

// 分块上传
const Upload_Part_Info_Key = "Upload_Info"      // 分块上传信息rediskey前缀
const Upload_Part_Slice_Expire = time.Hour * 24 // 分块上传文件分块保存时间，用于断点续传
const Upload_Part_Info_Hash_Key = "FileHash"    // 分块上传info map相关key
const Upload_Part_Info_Size_Key = "FileSize"
const Upload_Part_Info_ID_Key = "UploadID"
const Upload_Part_Info_CSize_Key = "ChunkSize"
const Upload_Part_Info_CCount_Key = "ChunkCount"
const Upload_Part_File_Info_Key = "File_Info"
const Uploac_Part_Info_Fileds = 6

// mq相关
const Routing_Key = "cos"
const Exchange = "cheetah_netdesk"
const Transfer_COS_Queue = "cheetah_netdesk_trans_cos"
const Transfer_COS_Queue_Err = "cheetah_netdesk_trans_cos_err"
const Default_Content_Type = "text/plain"

// 文件存储
const Default_System_Prefix = "test"         // 文件系统根目录，未来替换为root
const Default_System_parent = 0              // root对应ID
const Default_Sign_Expire = time.Minute * 15 // 签名默认有效时间
const Store_Type_COS = 0                     // 存储类型
const Store_Type_Tmp = 1                     // 临时存储
const Store_Type_Local = 2

// 文件分享
const Administrator_Uuid = "0"

const Default_Page_Size = 10
