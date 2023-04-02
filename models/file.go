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
	Path      string // 文件路径，即COS中的唯一KEY
	Size      int    // 文件大小
	Link      int    `json:"-"` // 文件引用数
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
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

type UploadObjectParams struct {
	// fileKey hash size file_uuid user_file_uuid Parent_Id User_Uuid
	FileKey        string
	Hash           string
	Size           int
	Parent         string
	User_Uuid      string
	File_Uuid      string // 可选
	User_File_Uuid string // 可选
}
