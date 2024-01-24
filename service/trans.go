package service

import (
	"NetDisk/client"
	"NetDisk/conf"
	"NetDisk/helper"
	"NetDisk/models"
	"strconv"

	"github.com/pkg/errors"
)

// TODO 优化为redis过期通知
func GetTransList(UserUuid string, pageNum, mod, status int) ([]*models.TransShow, error) {
	trans, err := client.GetDBClient().
		GetTransListByUser(UserUuid, pageNum, conf.DefaultPageSize, mod, status)
	if err != nil {
		return nil, err
	}
	// 读取时判断redis是否过期，过期则更改状态
	res := make([]*models.TransShow, len(trans))
	for i, t := range trans {
		key := ""
		if mod == conf.UploadMod {
			key = helper.GenUploadPartInfoKey(t.Uuid)
		} else {
			key = helper.GenDownloadPartInfoKey(t.Uuid)
		}

		curSize, chunkNum, chunkSize, chunkList := 0, 0, 0, []int{}
		// 如果为process
		if t.Status == conf.TransProcess {
			// 查看redis中是否过期
			flag, err := client.GetCacheClient().Exists(key)
			if err != nil {
				return nil, err
			}
			if flag == 0 {
				err = client.GetDBClient().UpdateTransState(t.Uuid, conf.TransFail)
				if err != nil {
					return nil, err
				}
				t.Status = conf.TransFail
			} else {
				// 没过期则读取配置
				infoMap, err := client.GetCacheClient().HGetAll(key)
				if err != nil {
					return nil, errors.Wrap(err, "[GetTransList] get trans info error: ")
				}
				if _, ok := infoMap[conf.UploadPartInfoCCountKey]; !ok {
					return nil, errors.Wrap(conf.MapNotHasError, "[GetTransList] get chunk count error: ")
				}
				// 分片总数
				chunkNum, _ = strconv.Atoi(infoMap[conf.UploadPartInfoCCountKey])
				chunkSize, _ = strconv.Atoi(infoMap[conf.UploadPartInfoCSizeKey])
				// 已传输分片数
				curNum := len(infoMap) - conf.UploadPartInfoFileds
				curSize = curNum * chunkSize
				// 分片列表
				for k, v := range infoMap {
					// 如果key是数字，说明时一个分片
					if k[0]-'0' >= 0 && k[0]-'0' < 10 {
						num, _ := strconv.Atoi(v)
						chunkList = append(chunkList, num)
					}
				}
			} // else
		} // if
		// 整合为前端需要的数据类型
		show := &models.TransShow{
			Uuid:       t.Uuid,
			FileUuid:   t.UserFileUuid, // 前端认为UserFileUuid是FileUuid
			UserUuid:   t.UserUuid,
			FileKey:    t.FileKey,
			LocalPath:  t.LocalPath,
			RemotePath: t.RemotePath,
			ParentUuid: t.ParentUuid,
			Hash:       t.Hash,
			Size:       t.Size,
			Name:       t.Name,
			Ext:        t.Ext,
			Status:     t.Status,
			Isdown:     t.Isdown,

			CurSize:    curSize,
			ChunkSize:  chunkSize,
			ChunkCount: chunkNum,
			ChunkList:  chunkList,
		}
		// 加入res
		res[i] = show
	}

	return res, nil
}
