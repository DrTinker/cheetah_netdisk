package models

type Trans struct {
	ID             uint   `gorm:"primaryKey" json:"-"`
	Uuid           string // uploadID / downloadID
	User_Uuid      string // 用户uuid
	User_File_Uuid string // user_file表中的uuid
	File_Uuid      string // file表中的uuid，用于索引，前端不可见
	File_Key       string // 文件cos key
	Local_Path     string // 文件用户本地路径
	Hash           string // 文件hash
	Size           int    // 文件大小
	Parent_Uuid    string // 父级目录uuid
	Name           string // 名 无后缀
	Ext            string // 后缀
	Status         int    // 当前上传状态 0: 上传中 1: 上传成功 2: 上传失败
}
