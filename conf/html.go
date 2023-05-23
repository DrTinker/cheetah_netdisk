package conf

// 保存HTML以及同前端交互的关键定义
// 邮件发送验证码页面HTML
const Email_Verify_Page = "<h2>您的验证码为：</h2><br><h1>%s</h1>"

const Page_Num_Key = "page"
const File_Ext_Key = "ext"

// 文件上传相关
const File_Form_Key = "files"
const File_Hash_Key = "md5"
const File_Uuid_Key = "file_uuid"
const File_Name_Key = "name"
const File_Size_Key = "size"
const Folder_Path_Key = "path"
const File_Path_Key = "path"
const Folder_Uuid_Key = "parent_uuid"
const File_Parent_Key = "parent"
const File_Src_Key = "src"            // 文件复制原地址
const File_Des_Key = "des"            // 文件复制目的地址
const File_Quick_Upload_Key = "quick" // 文件秒传标志
const Task_List_Key = "list"          // 批量操作

// 分块上传
const File_Chunk_Num_Key = "chunk_num"
const File_Upload_ID_Key = "upload_id"

// 分享
const Share_User_File_Uuid = "file_uuid"
const Share_Uuid = "share_uuid"
const Share_Code = "code"
const Share_Expire_Time = "expire_at"
