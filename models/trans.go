package models

type Trans struct {
	ID           uint   `gorm:"primaryKey" json:"-"`
	Uuid         string `json:"transID"`    // uploadID / downloadID
	UserUuid     string `json:"userID"`     // 用户uuid
	UserFileUuid string `json:"fileID"`     // user_file表中的uuid
	FileKey      string `json:"fileKey"`    // 文件cos key
	LocalPath    string `json:"localPath"`  // 文件用户本地路径
	RemotePath   string `json:"remotePath"` // 文件云空间存储路径
	Hash         string `json:"hash"`       // 文件hash
	Size         int    `json:"size"`       // 文件大小
	ParentUuid   string `json:"parent"`     // 父级目录uuid
	Name         string `json:"name"`       // 名 无后缀
	Ext          string `json:"ext"`        // 后缀
	Status       int    `json:"status"`     // 当前上传状态 0: 上传中 1: 上传成功 2: 上传失败
	Isdown       int    `json:"isdown"`     // 0为上传 1为下载
}

type TransShow struct {
	Uuid       string `json:"transID"`    // uploadID / downloadID
	UserUuid   string `json:"userID"`     // 用户uuid
	FileUuid   string `json:"fileID"`     // user_file表中的uuid
	FileKey    string `json:"fileKey"`    // 文件cos key
	LocalPath  string `json:"localPath"`  // 文件用户本地路径
	RemotePath string `json:"remotePath"` // 文件云空间存储路径
	Hash       string `json:"hash"`       // 文件hash
	Size       int    `json:"size"`       // 文件大小
	ParentUuid string `json:"parent"`     // 父级目录uuid
	Name       string `json:"name"`       // 名 无后缀
	Ext        string `json:"ext"`        // 后缀
	Status     int    `json:"status"`     // 当前上传状态 0: 上传中 1: 上传成功 2: 上传失败
	Isdown     int    `json:"isdown"`     // 0为上传 1为下载

	// 下面的只有process状态的才有
	CurSize    int   `json:"curSize"` // 已上传大小
	ChunkSize  int   `json:"chunkSize"`
	ChunkCount int   `json:"chunkCount"`
	ChunkList  []int `json:"chunkList"` // 已上传的分块列表
}
