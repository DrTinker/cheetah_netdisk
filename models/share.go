package models

import (
	"database/sql"
	"time"
)

type Share struct {
	ID             uint         `gorm:"primaryKey" json:"-"`
	Uuid           string       // 前端可见的分享id
	User_Uuid      string       // 用户uuid
	User_File_Uuid string       // user_file表中的uuid
	File_Uuid      string       `json:"-"` // file表中的uuid，用于索引，前端不可见
	Fullname       string       // 文件名
	Code           string       // 分享密码
	Expire_Time    sql.NullTime `gorm:"type:TIMESTAMP NULL"` // 有效时间
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ShareShow struct {
	Uuid           string // 前端可见的分享id
	User_Uuid      string // 用户uuid
	User_File_Uuid string // user_file表中的uuid
	Fullname       string // 文件名
	Code           string // 分享密码
	Status         int    // 1: 有效 2: 过期
	Expire_Time    string // 有效时间
	CreatedAt      string
	UpdatedAt      string
}
