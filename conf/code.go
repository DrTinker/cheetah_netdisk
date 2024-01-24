package conf

// RPC状态码
const RPC_SUCCESS_CODE = 10000
const RPC_FAILED_CODE = 10001

// HTTP状态码
// 成功
const HTTP_SUCCESS_CODE = 10000

// 失败
const SERVER_ERROR_CODE = 10001 // 服务器内部错误

// 错误
// 登录
const HTTP_INVALID_PARAMS_CODE = 20000 // 参数错误通用
const ERROR_LOGIN_CODE = 21000
const ERROR_REGISTER_CODE = 21001
const ERROR_AUTH_CHECK_TOKEN_FAIL_CODE = 21002    // jwt
const ERROR_AUTH_CHECK_TOKEN_TIMEOUT_CODE = 21003 // jwt
const ERROR_VERIFY_CODE = 21004                   // 验证码生成错误

// 邮件
const ERROR_EMAIL_SEND_CODE = 21005

// 传输
const ERROR_UPLOAD_CODE = 22000
const ERROR_VOLUME_COUNT_CODE = 22001
const ERROR_FILE_CHECK_CODE = 22002 // 文件校验时出错
const ERROR_FILE_EXIST_CODE = 22003 // 文件存在
const QUICK_UPLOAD_CODE = 22004     // 文件存在
const ERROR_FILE_HASH_CODE = 22005  // 文件md5值无效
const ERROR_GET_URL_CODE = 22006    // 获取预签名错误
const FILE_EXIST_CODE = 22007       // 同一个用户上传相同文件
const ERROR_FILE_OWNER_CODE = 22009 // 下载文件的用户不是文件拥有者
// 分块
const CHUNK_MISS_CODE = 22008 // 分片传输不完整

// 获取文件列表
const ERROR_INVAILD_PAGE_CODE = 22005 // 给定页数超过最大页数

// 文件系统
const ERROR_LIST_FILES_CODE = 23000
const ERROR_GET_INFO_CODE = 23001
const ERROR_FILE_COPY_CODE = 23002
const ERROR_UPDATE_NAME_CODE = 23003
const ERROR_DELETE_FILE_CODE = 23004

// 分享
const ERROR_CREATE_ShareCode = 24000
const WARN_SHARE_EXPIRES_CODE = 24001
const RECORD_DELETED_CODE = 24002
