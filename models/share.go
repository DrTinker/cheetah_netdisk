package models

import (
	"time"

	"gorm.io/gorm"
)

type Share struct {
	ID             uint   `gorm:"primaryKey" json:"-"`
	Uuid           string // 前端可见的分享id
	User_Uuid      string // 用户uuid
	User_File_Uuid string // user_file表中的uuid
	File_Uuid      string // file表中的uuid，用于索引，前端不可见
	Expire         int    // 有效时间
	Click          int    // 点击数
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}
