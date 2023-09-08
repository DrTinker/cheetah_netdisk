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
	UserUUID string `json:"userID" form:"userID"`
	Password string `json:"userPwd" form:"userPwd"`
	Email    string `json:"userEmail" form:"userEmail"`
}

type User struct {
	ID          uint           `gorm:"primaryKey" json:"-"`
	Uuid        string         `json:"userID"`
	Name        string         `json:"userName"`
	Password    string         `json:"userPwd"`
	Email       string         `json:"userEmail"`
	Phone       string         `json:"userPhone"`
	Level       int            `json:"level"`       // 0: 普通用户， 1：VIP用户， 2：特权用户
	StartUuid   string         `json:"startID"`     // 用户文件空间根目录uuid，对应user_file表的uuid
	NowVolume   int64          `json:"nowVolume"`   // 已使用存储容量，单位B
	TotalVolume int64          `json:"totalVolume"` // 总存储容量，单位B
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type UserInfo struct {
	Uuid  string
	Name  string
	Level int // 0: 普通用户， 1：VIP用户， 2：特权用户
}
