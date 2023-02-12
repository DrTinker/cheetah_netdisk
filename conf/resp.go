package conf

const SUCCESS_RESP_MESSAGE = "Success!"

// 参数错误
const HTTP_INVALID_PARAMS_MESSAGE = "invaild params"

// 登录
const LOGIN_ERROR_MESSAGE = "ID or pwd error!"
const JWT_ERROR_MESSAGE = "Token parse error!"
const VERIFY_CODE_GEN_ERROR_MESSAGE = "Verify code generate error!"
const VERIFY_CODE_ERROR_MESSAGE = "Verify code invaild"
const REGISTER_ERROR_MESSAGE = "Register error!"
const REGISTER_REPEAT_MESSAGE = "User exist!"

// 上传
const UPLOAD_FAIL_MESSAGE = "File upload error: %s"
const GET_QUERY_ERROR_MESSAGE = ""
const GET_VOLUME_ERROR_MESSAGE = "Get user volume error"
const VOLUME_RUNOUT_ERROR_MESSAGE = "User volume run out"
const FILE_CHECK_ERROR_MESSAGE = "File check error"
const FILE_EXIST_MESSAGE = "File already exist"
const UPLOAD_SUCCESS_MESSAGE = "File upload success"
const UPLOAD_PART_INIT_FAIL_MESSAGE = "Init mutipart upload error"
const FILE_HASH_INVAILD_MESSAGE = "File md5 value invaild"

// 邮件
const SUCCESS_EMAIL_MESSAGE = "Verify code send success"
const FAIL_EMAIL_MESSAGE = "Verify code send fail"

// 文件系统
const LIST_FILES_FAIL_MESSAGE = "Get file list error"
const LIST_FILES_SUCCESS_MESSAGE = "Get file list success"
const GET_INFO_FAIL_MESSAGE = "Get file info error"
const COPY_FILE_FAIL_MESSAGE = "Copy file error"
