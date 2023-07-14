package models

import "time"

type MultiFileUploadOptions struct {
	ThreadPoolSize int // 使用线程数
	PartSize       int
	CheckPoint     bool // 使用断点续传
}

// service层UploadFile参数
type TransObjectParams struct {
	// fileKey hash size file_uuid user_file_uuid Parent_Id User_Uuid
	UploadID       string
	FileKey        string
	LocalPath      string // 用户本地存储路径
	Hash           string
	Size           int
	Parent         string
	User_Uuid      string
	Name           string
	Ext            string
	File_Uuid      string // 可选
	User_File_Uuid string // 可选
}

// 初始化分片上传返回值
type InitTransResult struct {
	UploadID   string
	Quick      bool  // 秒传标志
	ChunkCount int   // 总计分片数
	ChunkList  []int // 断点续传，已经上传的分片列表
}

// 创建分享链接参数
type CreateShareParams struct {
	Share_Uuid     string
	User_Uuid      string
	User_File_Uuid string
	Code           string
	Expire         time.Time
}
