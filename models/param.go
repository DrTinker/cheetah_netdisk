package models

type MultiFileUploadOptions struct {
	ThreadPoolSize int // 使用线程数
	PartSize       int
	CheckPoint     bool // 使用断点续传
}

// service层UploadObjectServer参数
type UploadObjectParams struct {
	// fileKey hash size file_uuid user_file_uuid Parent_Id User_Uuid
	FileKey        string
	Hash           string
	Size           int
	Parent         string
	User_Uuid      string
	Name           string
	Ext            string
	File_Uuid      string // 可选
	User_File_Uuid string // 可选
}
