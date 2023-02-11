package conf

import "time"

// HTTP启动项
const App = "conf/app.ini"
const HttpServer = "HttpServer"

// jwt
const JWTKeyValue = "cheetah_net_desk"
const JWTExpireValue = 36000
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

// 邮件
const Email_Verify_MSG = "Cheetah Net Desk验证码"
const Email_Verify_Name = "cheetah_net_desk"

// 验证码
const Code_Cache_Key = "verify_code_key"
const Code_Len = 6
const Code_Expire = 5 * time.Minute
const Code_Param_Key = "code"

// 文件上传
const File_Part_Size_Max = 1024 * 1024 * 16 // 1MB
const Default_Thread_Pool_Size = 5
const Folder_Default_Size = 1
const Folder_Default_EXT = "folder"
const File_Exist_Flag = "exist"

// 文件存储
const Default_System_Prefix = "test" // 文件系统根目录，未来替换为root
