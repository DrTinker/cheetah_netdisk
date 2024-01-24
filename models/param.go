package models

import (
	"database/sql"
	"net/http"
)

type MultiFileUploadOptions struct {
	ThreadPoolSize int // 使用线程数
	PartSize       int
	CheckPoint     bool // 使用断点续传
}

// service层UploadFile参数
type UploadObjectParams struct {
	// fileKey hash size FileUuid UserFileUuid ParentId UserUuid
	UploadID     string
	FileKey      string
	LocalPath    string // 用户本地存储路径
	RemotePath   string // 云空间存储路径
	Hash         string
	Size         int
	Parent       string
	UserUuid     string
	Name         string
	Ext          string
	FileUuid     string // 可选
	UserFileUuid string // 可选
}

// 初始化分片上传返回值
type InitUploadResult struct {
	UploadID   string
	Quick      bool  // 秒传标志
	ChunkCount int   // 总计分片数
	ChunkList  []int // 断点续传，已经上传的分片列表
}

// 初始化分片下载返回值
type InitDownloadResult struct {
	DownloadID string
	ChunkCount int    // 总计分片数
	ChunkList  []int  // 断点续传，已经上传的分片列表
	Hash       string // 文件hash，用于客户端合并后检查文件
	Url        string // cos访问签名
}

// 分片下载
type DownloadObjectParam struct {
	Req          http.Request
	Resp         http.ResponseWriter
	DownloadID   string
	UserFileUuid string
	UserUuid     string
	ParentUuid   string // 文件所在目录uuid
	LocalPath    string // 用户本地存储路径
	RemotePath   string // 云存储路径
	Continue     bool   // 是否续传
}

// 创建分享链接参数
type CreateShareParams struct {
	ShareUuid    string
	UserUuid     string
	UserFileUuid string
	Code         string
	Fullname     string
	Expire       sql.NullTime
}
