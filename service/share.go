package service

import (
	"NetDesk/client"
	"NetDesk/models"
	"time"

	"github.com/sirupsen/logrus"
)

// 创建链接
func CreateShareLink(param *models.CreateShareParams) error {
	// 解析参数
	share_uuid := param.Share_Uuid
	user_uuid := param.User_Uuid
	user_file_uuid := param.User_File_Uuid
	expire := param.Expire
	code := param.Code
	// 通过user_file_uuid查询file_uuid
	file_uuid, err := client.GetDBClient().GetFileUuidByUserFileUuid(user_file_uuid)
	if err != nil {
		return err
	}
	// 封装结构体
	shareInfo := &models.Share{
		Uuid:           share_uuid,
		User_Uuid:      user_uuid,
		User_File_Uuid: user_file_uuid,
		File_Uuid:      file_uuid,
		Code:           code,
		Expire_Time:    expire,
		Click_Num:      0,
	}
	// 存数据库
	err = client.GetDBClient().CreateShare(shareInfo)
	if err != nil {
		return err
	}

	return nil
}

// 查询share
func GetShareInfo(share_uuid string) (info *models.Share, time_out bool, err error) {
	// 通过uuid查询share信息
	info, err = client.GetDBClient().GetShareByUuid(share_uuid)
	if err != nil {
		return nil, false, err
	}
	// 检查过期时间
	now := time.Now()
	if info.Expire_Time.Before(now) {
		return info, true, nil
	}
	// 增加点击数，不像上层传递错误
	err = client.GetDBClient().UpdateClickNumByUuid(share_uuid)
	if err != nil {
		logrus.Warn("[GetShareInfo] increase click err: ", share_uuid)
	}
	// 未过期返回
	return info, false, nil
}

// 通过分享获取文件
func CopyFileByShare(share_uuid, des_uuid, user_uuid string) error {
	// 通过share_uuid获取user_file_uuid
	src_uuid, err := client.GetDBClient().GetUserFileUuidByShareUuid(share_uuid)
	if err != nil {
		return err
	}
	err = CopyObject(src_uuid, des_uuid, user_uuid)
	if err != nil {
		return err
	}
	return nil
}

// 取消分享
func CancelShare(share_uuid string) error {
	err := client.GetDBClient().DeleteShareByUuid(share_uuid)
	if err != nil {
		return err
	}

	return nil
}
