package models

type Trans struct {
	ID             uint   `gorm:"primaryKey" json:"-"`
	Uuid           string // uploadID / downloadID
	User_Uuid      string // 用户uuid
	User_File_Uuid string // user_file表中的uuid
	File_Key       string // 文件cos key
	Local_Path     string // 文件用户本地路径
	Remote_Path    string // 文件云空间存储路径
	Hash           string // 文件hash
	Size           int    // 文件大小
	Parent_Uuid    string // 父级目录uuid
	Name           string // 名 无后缀
	Ext            string // 后缀
	Status         int    // 当前上传状态 0: 上传中 1: 上传成功 2: 上传失败
	Isdown         int    // 0为上传 1为下载
}

type TransShow struct {
	Uuid        string // uploadID / downloadID
	User_Uuid   string // 用户uuid
	File_Uuid   string // user_file表中的uuid
	File_Key    string // 文件cos key
	Local_Path  string // 文件用户本地路径
	Remote_Path string // 文件云空间存储路径
	Hash        string // 文件hash
	Size        int    // 文件大小
	Parent_Uuid string // 父级目录uuid
	Name        string // 名 无后缀
	Ext         string // 后缀
	Status      int    // 当前上传状态 0: 上传中 1: 上传成功 2: 上传失败
	Isdown      int    // 0为上传 1为下载

	// 下面的只有process状态的才有
	CurSize    int // 已上传大小
	ChunkSize  int
	ChunkCount int
	ChunkList  []int // 已上传的分块列表
}
