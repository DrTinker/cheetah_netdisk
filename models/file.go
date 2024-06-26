package models

import (
	"time"

	"gorm.io/gorm"
)

type File struct { // file中的一条记录唯一对应一个COS中的实际文件，可对应多条userfile中的记录，复制出来的文件COS中不进行额外存储
	ID        uint   `gorm:"primaryKey" json:"-"`
	Uuid      string // 前端不可见
	Name      string // 文件名称
	Hash      string // 哈希值判断文件存在性
	Ext       string // 文件扩展名
	FileKey   string // 文件路径，即COS中的唯一KEY, 为test/hash.ext(测试阶段)或root/hash.ext(正式阶段)
	Thumbnail string // 文件缩略图存储路径
	Size      int    // 文件大小
	Link      int    `json:"-"` // 文件引用数
	StoreType int    // 存储类型 0: cos 1: tmp 2: local
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UserFile struct { // userfile中的一条记录唯一对应用户存储空间一个文件，包括复制出来的文件
	ID        uint   `gorm:"primaryKey" json:"-"`
	Uuid      string // 前端可见的文件id
	UserUuid  string // 用户uuid
	ParentId  int    `json:"-"` // 父节点id，id为user_file表的id
	FileUuid  string `json:"-"` // file表中的uuid，用于索引，前端不可见
	Ext       string // 文件扩展名
	Name      string // 文件名称
	Size      int    // 文件大小
	Thumbnail string // 文件缩略图存储路径
	Hash      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserFileShow struct {
	Uuid      string `json:"fileID"`    // 前端可见的文件id
	UserUuid  string `json:"userID"`    // 用户uuid
	Ext       string `json:"ext"`       // 文件扩展名
	Name      string `json:"name"`      // 文件名称
	Size      int    `json:"size"`      // 文件大小
	Thumbnail string `json:"thumbnail"` // 缩略图地址
	Hash      string `json:"hash"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type Part struct {
	ETag    string
	PartNum int
}

// 文件异步上传cos相关结构体
type TransferSetting struct {
	Exchange  string
	RoutinKey string
}

// 消息队列消息结构
type TransferMsg struct {
	TransID   string // 上传ID唯一
	FileHash  string
	FileName  string // 文件全名
	TnName    string // 缩略图全名
	TmpPath   string // 私有云存储路径，默认与fileKey同
	FileKey   string // cos filekey
	Thumbnail string // 缩略图私有云存储路径，默认与TnFileKey同
	TnFileKey string // 缩略图存储fileKey
	StoreType int    // 0：cos 1：本地
	Task      int    // 0: 上传 1: 下载
}

// 分块上传结构体
type UploadPartInfo struct {
	FileHash   string // 文件哈希值
	FileSize   int    // 文件总大小
	UploadID   string // 上传ID唯一
	ChunkSize  int    // 分块大小
	ChunkCount int    // 分块数量
}

// 批量操作参数
type BatchTaskInfo struct {
	Des string   `json:"des"` // 目标文件夹uuid
	Src []string `json:"src"` // 要操作文件uuid列表
}

type MediaFilter struct {
	PicFilter map[string]bool //{"jpg", "jpeg", "png", "gif"}
	// FLV 、AVI、MOV、MP4、WMV
	VideoFilter map[string]bool //{"mp4", "flv", "avi", "mov", "wmv"}
	// MP3，WMA，WAV，APE，FLAC，OGG，AAC
	AideoFilter map[string]bool //{"mp3", "wma", "wav", "ape", "flac", "ogg", "aac"}
	// rar、zip、arj、tar
	PackFilter map[string]bool //{"rar", "zip", "arj", "tar", "gz"}
	// execel ppt doc docx md txt
	DocFilter map[string]bool //{"execel", "ppt", "doc", "docx", "md", "txt"}
}
