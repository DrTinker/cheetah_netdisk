package models

import (
	"database/sql"
	"time"
)

type Share struct {
	ID           uint         `gorm:"primaryKey" json:"-"`
	Uuid         string       // 前端可见的分享id
	UserUuid     string       // 用户uuid
	UserFileUuid string       // user_file表中的uuid
	FileUuid     string       `json:"-"` // file表中的uuid，用于索引，前端不可见
	Fullname     string       // 文件名
	Code         string       // 分享密码
	ExpireTime   sql.NullTime `gorm:"type:TIMESTAMP NULL"` // 有效时间
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ShareShow struct {
	Uuid         string `json:"shareID"`    // 前端可见的分享id
	UserUuid     string `json:"userID"`     // 用户uuid
	UserFileUuid string `json:"fileID"`     // user_file表中的uuid
	Fullname     string `json:"fullname"`   // 文件名
	Code         string `json:"code"`       // 分享密码
	Status       int    `json:"status"`     // 1: 有效 2: 过期
	ExpireTime   string `json:"expireTime"` // 有效时间
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}
