package models

import (
	"time"

	"gorm.io/gorm"
)

type Token struct {
	ID       string `json:"user_ID"`
	Email    string `json:"user_email"`
	Password string `json:"user_pwd"`
	Expire   int64  `json:"expire"`
}

type Login struct {
	User_UUID string `json:"user_id" form:"user_id"`
	Password  string `json:"user_pwd" form:"user_pwd"`
	Email     string `json:"user_email" form:"user_email"`
}

type User struct {
	ID           uint `gorm:"primaryKey" json:"-"`
	Uuid         string
	Name         string
	Password     string `json:"password,omitempty"`
	Email        string
	Phone        string
	Level        int    // 0: 普通用户， 1：VIP用户， 2：特权用户
	Start_Uuid   string // 用户文件空间根目录uuid，对应user_file表的uuid
	Now_Volume   int64  // 已使用存储容量，单位B
	Total_Volume int64  // 总存储容量，单位B
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type UserInfo struct {
	Uuid  string
	Name  string
	Level int // 0: 普通用户， 1：VIP用户， 2：特权用户
}
