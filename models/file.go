package models

import (
	"time"

	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

type File struct { // file中的一条记录唯一对应一个COS中的实际文件，可对应多条userfile中的记录，复制出来的文件COS中不进行额外存储
	ID         uint   `gorm:"primaryKey" json:"-"`
	Uuid       string // 前端不可见
	Name       string // 文件名称
	Hash       string // 哈希值判断文件存在性
	Ext        string // 文件扩展名
	File_Key   string // 文件路径，即COS中的唯一KEY, 为test/hash.ext(测试阶段)或root/hash.ext(正式阶段)
	Size       int    // 文件大小
	Link       int    `json:"-"` // 文件引用数
	Store_Type int    // 存储类型 0: cos 1: tmp 2: local
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type UserFile struct { // userfile中的一条记录唯一对应用户存储空间一个文件，包括复制出来的文件
	ID        uint   `gorm:"primaryKey" json:"-"`
	Uuid      string // 前端可见的文件id
	User_Uuid string // 用户uuid
	Parent_Id int    `json:"-"` // 父节点id，id为user_file表的id
	File_Uuid string `json:"-"` // file表中的uuid，用于索引，前端不可见
	Ext       string // 文件扩展名
	Name      string // 文件名称
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserFileShow struct {
	Uuid      string // 前端可见的文件id
	User_Uuid string // 用户uuid
	Ext       string // 文件扩展名
	Name      string // 文件名称
	CreatedAt string
	UpdatedAt string
}

type Part struct {
	ETag    string
	PartNum int
}

// 文件异步上传cos相关结构体
type TransferSetting struct {
	Channel   *amqp.Channel
	Exchange  string
	RoutinKey string
}

// 消息队列消息结构
type TransferMsg struct {
	FileHash  string
	Src       string // 本地存储路径
	Des       string // cos filekey
	StoreType int    // 0：cos 1：本地
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
	Des string   // 目标文件夹uuid
	Src []string // 要操作文件uuid列表
}
