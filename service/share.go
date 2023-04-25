package service

import (
	"NetDesk/client"
	"NetDesk/models"
	"time"

	"github.com/pkg/errors"
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
		return errors.Wrap(err, "[CreateShareLink] get file uuid error: ")
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
		return errors.Wrap(err, "[CreateShareLink] create share record error: ")
	}

	return nil
}

// 查询share
func GetShareInfo(share_uuid string) (info *models.Share, time_out bool, err error) {
	// 通过uuid查询share信息
	info, err = client.GetDBClient().GetShareByUuid(share_uuid)
	if err != nil {
		return nil, false, errors.Wrap(err, "[GetShareInfo] get share info error: ")
	}
	// 检查过期时间
	now := time.Now()
	if info.Expire_Time.Before(now) {
		return info, true, nil
	}
	// 未过期返回
	return info, false, nil
}

// 取消分享
func CancelShare(share_uuid string) error {
	err := client.GetDBClient().DeleteShareByUuid(share_uuid)
	if err != nil {
		return errors.Wrap(err, "[CancelShare] delete share record error: ")
	}

	return nil
}
