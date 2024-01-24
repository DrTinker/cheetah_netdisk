package conf

// 保存HTML以及同前端交互的关键定义
// 邮件发送验证码页面HTML
const EmailVerifyPage = "<h2>您的验证码为：</h2><br><h1>%s</h1>"
const ForgetPasswordPage = "<h2>您的密码为：</h2><br><h1>%s</h1>"

const UserID = "userID"
const UserPWD = "userPwd"
const UserEmail = "userEmail"
const UserPhone = "userPhone"
const UserName = "userName"

const PageNumKey = "page"
const FileExtKey = "ext"

// 文件上传相关
const FileFormKey = "files"
const FileHashKey = "hash"
const FileUuidKey = "fileID"
const FileNameKey = "name"
const FileLocalPathKey = "localPath"
const FileRemotePathKey = "remotePath"
const FileSizeKey = "size"
const FolderPathKey = "path"
const FilePathKey = "fileKey"
const FileParentKey = "parent"
const FileSrcKey = "src"           // 文件复制原地址
const FileDesKey = "des"           // 文件复制目的地址
const FileQuickUploadKey = "quick" // 文件秒传标志
const TaskListKey = "list"         // 批量操作

// 分块上传
const FileChunkNumKey = "chunkNum"
const FileUploadIDKey = "uploadID"

// 分块下载
const FileDownloadIDKey = "downloadID"

// 分享
const ShareUserFileUuid = "fileID"
const ShareUuid = "shareID"
const ShareUserUuid = "userID"
const ShareCode = "code"
const ShareName = "fullname"
const ShareMod = "mod"
const ShareExpireTime = "expireAt"

// 传输
const TransUuidKey = "transID"
const TransIsdownKey = "isdown"
const TransStatusKey = "status"
