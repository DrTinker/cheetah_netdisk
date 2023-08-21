package service

import (
	"NetDesk/client"
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
	"time"
)

// 创建链接
func SetShareLink(param *models.CreateShareParams) error {
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
		Fullname:       param.Fullname,
		Code:           code,
		Expire_Time:    expire,
	}
	// 存数据库
	err = client.GetDBClient().SetShare(shareInfo)
	if err != nil {
		return err
	}

	return nil
}

// 查询share
func GetShareInfo(share_uuid string) (res *models.ShareShow, time_out bool, err error) {
	// 通过uuid查询share信息
	info, err := client.GetDBClient().GetShareByUuid(share_uuid)
	if err != nil {
		return nil, false, err
	}
	// 查看文件是否被删除
	flag, err := client.GetDBClient().GetUserFileByUuid(info.User_File_Uuid)
	if err != nil {
		return nil, false, err
	}
	if flag == nil {
		return nil, true, conf.FileDeletedError
	}
	res = &models.ShareShow{
		Uuid:           info.Uuid,
		User_Uuid:      info.User_Uuid,
		User_File_Uuid: info.User_File_Uuid,
		Code:           info.Code,
		Fullname:       info.Fullname,
		Status:         conf.Share_Expire_Mod,
		Expire_Time:    helper.TimeFormat(info.Expire_Time.Time),
		CreatedAt:      helper.TimeFormat(info.CreatedAt),
		UpdatedAt:      helper.TimeFormat(info.UpdatedAt),
	}
	// 没有过期时间视为永久有效
	if !info.Expire_Time.Valid {
		res.Expire_Time = ""
	}
	// 检查过期时间
	now := time.Now()
	// 有过期时间且过期时间在当前时间之前
	if info.Expire_Time.Valid && info.Expire_Time.Time.Before(now) {
		res.Status = conf.Share_Out_Mod
		return res, true, nil
	}
	// 未过期返回
	return res, false, nil
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

// 取消分享
func CancelBatchShare(cancelList []string) error {
	for _, c := range cancelList {
		err := client.GetDBClient().DeleteShareByUuid(c)
		if err != nil {
			return err
		}
	}
	return nil
}

// 获取用户分享列表
func GetShareList(user_uuid string, cur, mod int) ([]*models.ShareShow, error) {
	infos, err := client.GetDBClient().GetShareListByUser(user_uuid, cur, conf.Default_Page_Size, mod)
	if err != nil {
		return nil, err
	}
	res := make([]*models.ShareShow, len(infos))
	for i, info := range infos {
		tmp := &models.ShareShow{
			Uuid:           info.Uuid,
			User_Uuid:      info.User_Uuid,
			User_File_Uuid: info.User_File_Uuid,
			Code:           info.Code,
			Fullname:       info.Fullname,
			Status:         conf.Share_Expire_Mod,
			Expire_Time:    helper.TimeFormat(info.Expire_Time.Time),
			CreatedAt:      helper.TimeFormat(info.CreatedAt),
			UpdatedAt:      helper.TimeFormat(info.UpdatedAt),
		}
		// 没有过期时间视为永久有效
		if !info.Expire_Time.Valid {
			tmp.Expire_Time = ""
		}
		// 检查过期时间
		now := time.Now()
		// 有过期时间且过期时间在当前时间之前
		if info.Expire_Time.Valid && info.Expire_Time.Time.Before(now) {
			tmp.Status = conf.Share_Out_Mod
		}
		res[i] = tmp
	}
	return res, nil
}
