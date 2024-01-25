package service

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"time"
)

// 创建链接
func SetShareLink(param *models.CreateShareParams) error {
	// 解析参数
	ShareUuid := param.ShareUuid
	UserUuid := param.UserUuid
	UserFileUuid := param.UserFileUuid
	expire := param.Expire
	code := param.Code
	// 通过UserFileUuid查询FileUuid
	FileUuid, err := client.GetDBClient().GetFileUuidByUserFileUuid(UserFileUuid)
	if err != nil {
		return err
	}
	// 封装结构体
	shareInfo := &models.Share{
		Uuid:         ShareUuid,
		UserUuid:     UserUuid,
		UserFileUuid: UserFileUuid,
		FileUuid:     FileUuid,
		Fullname:     param.Fullname,
		Code:         code,
		ExpireTime:   expire,
	}
	// 存数据库
	err = client.GetDBClient().SetShare(shareInfo)
	if err != nil {
		return err
	}

	return nil
}

// 查询share
func GetShareInfo(ShareUuid string) (res *models.ShareShow, time_out bool, err error) {
	// 通过uuid查询share信息
	info, err := client.GetDBClient().GetShareByUuid(ShareUuid)
	if err != nil {
		return nil, false, err
	}
	// 查看文件是否被删除
	flag, err := client.GetDBClient().GetUserFileByUuid(info.UserFileUuid)
	if err != nil {
		return nil, false, err
	}
	if flag == nil {
		return nil, true, conf.FileDeletedError
	}
	res = &models.ShareShow{
		Uuid:         info.Uuid,
		UserUuid:     info.UserUuid,
		UserFileUuid: info.UserFileUuid,
		Code:         info.Code,
		Fullname:     info.Fullname,
		Status:       conf.ShareExpireMod,
		ExpireTime:   helper.TimeFormat(info.ExpireTime.Time),
		CreatedAt:    helper.TimeFormat(info.CreatedAt),
		UpdatedAt:    helper.TimeFormat(info.UpdatedAt),
	}
	// 没有过期时间视为永久有效
	if !info.ExpireTime.Valid {
		res.ExpireTime = ""
	}
	// 检查过期时间
	now := time.Now()
	// 有过期时间且过期时间在当前时间之前
	if info.ExpireTime.Valid && info.ExpireTime.Time.Before(now) {
		res.Status = conf.ShareOutMod
		return res, true, nil
	}
	// 未过期返回
	return res, false, nil
}

// 通过分享获取文件
func CopyFileByShare(ShareUuid, des_uuid, UserUuid string) error {
	// 通过ShareUuid获取UserFileUuid
	src_uuid, err := client.GetDBClient().GetUserFileUuidByShareUuid(ShareUuid)
	if err != nil {
		return err
	}
	err = CopyObject(src_uuid, des_uuid, UserUuid)
	if err != nil {
		return err
	}
	return nil
}

// 取消分享
func CancelShare(ShareUuid string) error {
	err := client.GetDBClient().DeleteShareByUuid(ShareUuid)
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
func GetShareList(UserUuid string, cur, mod int) ([]*models.ShareShow, error) {
	infos, err := client.GetDBClient().GetShareListByUser(UserUuid, cur, conf.DefaultPageSize, mod)
	if err != nil {
		return nil, err
	}
	res := make([]*models.ShareShow, len(infos))
	for i, info := range infos {
		tmp := &models.ShareShow{
			Uuid:         info.Uuid,
			UserUuid:     info.UserUuid,
			UserFileUuid: info.UserFileUuid,
			Code:         info.Code,
			Fullname:     info.Fullname,
			Status:       conf.ShareExpireMod,
			ExpireTime:   helper.TimeFormat(info.ExpireTime.Time),
			CreatedAt:    helper.TimeFormat(info.CreatedAt),
			UpdatedAt:    helper.TimeFormat(info.UpdatedAt),
		}
		// 没有过期时间视为永久有效
		if !info.ExpireTime.Valid {
			tmp.ExpireTime = ""
		}
		// 检查过期时间
		now := time.Now()
		// 有过期时间且过期时间在当前时间之前
		if info.ExpireTime.Valid && info.ExpireTime.Time.Before(now) {
			tmp.Status = conf.ShareOutMod
		}
		res[i] = tmp
	}
	return res, nil
}
