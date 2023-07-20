package service

import (
	"NetDesk/client"
	"NetDesk/conf"
	"NetDesk/helper"
	"NetDesk/models"
	"fmt"

	"github.com/sirupsen/logrus"
)

var filter = &models.MediaFilter{
	PicFilter: map[string]bool{"jpg": true, "jpeg": true, "png": true, "gif": true},
	//{"mp4", "flv", "avi", "mov", "wmv"}
	VideoFilter: map[string]bool{"mp4": true, "flv": true, "avi": true, "mov": true, "wmv": true},
	//{"mp3", "wma", "wav", "ape", "flac", "ogg", "aac"}
	AideoFilter: map[string]bool{"mp3": true, "wma": true, "wav": true, "ape": true, "flac": true, "ogg": true, "aac": true},
	//{"rar", "zip", "arj", "tar", "gz"}
	PackFilter: map[string]bool{"rar": true, "zip": true, "arj": true, "tar": true, "gz": true},
	// execel ppt doc docx md txt
	DocFilter: map[string]bool{"execel": true, "ppt": true, "doc": true, "docx": true, "md": true, "txt": true},
}

// path：原图像 or 视频地址，ext：扩展名
func MediaHandler(path, ext string) (tnPath, tnName string) {
	flag, err := helper.PathExists(path)
	if err != nil || !flag {
		logrus.Error("[MediaHandler] open media err or not exist: ", err)
	}
	for i := 0; i < 5; i++ {
		switch i {
		case 0:
			flag, tnPath, tnName := picHandler(path, ext)
			if flag && tnPath != "" {
				return tnPath, tnName
			}
		case 1:
			flag, tnPath, tnName := videoHandler(path, ext)
			if flag && tnPath != "" {
				return tnPath, tnName
			}
		}
	}

	return "", ""
}

func picHandler(path, ext string) (flag bool, tnPath, tnName string) {
	// 不是图片扩展名
	if _, ok := filter.PicFilter[ext]; !ok {
		return false, "", ""
	}
	// 获取tmp路径
	cfg, err := client.GetConfigClient().GetLocalConfig()
	if err != nil {
		logrus.Error("[MediaHandler] get loacl config error: ", err)
		return false, "", ""
	}
	// 调用client
	name, _, err := helper.SplitFilePath(path)
	if err != nil {
		logrus.Error("[MediaHandler] open media err: ", err)
		return false, "", ""
	}
	tnName = fmt.Sprintf("%s_tn.%s", name, conf.Default_Thumbnail_Ext)
	tnPath = cfg.TmpPath + "/" + tnName
	err = client.GetMediaClient().GetPicThumbNail(path, tnPath)
	if err != nil {
		logrus.Error("[MediaHandler] get thumbnail err: ", err)
		return false, "", ""
	}

	return true, tnPath, tnName
}

func videoHandler(path, ext string) (flag bool, tnPath, tnName string) {
	// 不是图片扩展名
	if _, ok := filter.VideoFilter[ext]; !ok {
		return false, "", ""
	}
	// 获取tmp路径
	cfg, err := client.GetConfigClient().GetLocalConfig()
	if err != nil {
		logrus.Error("[MediaHandler] get loacl config error: ", err)
		return false, "", ""
	}
	// 调用client
	name, _, err := helper.SplitFilePath(path)
	if err != nil {
		logrus.Error("[MediaHandler] open media err: ", err)
		return false, "", ""
	}
	tnName = fmt.Sprintf("%s_tn.%s", name, conf.Default_Thumbnail_Ext)
	tnPath = cfg.TmpPath + "/" + tnName
	err = client.GetMediaClient().GetVideoThumbNail(path, tnPath, conf.Default_ThumbNail_Frame)
	if err != nil {
		logrus.Error("[MediaHandler] get thumbnail err: ", err)
		return false, "", ""
	}

	return true, tnPath, tnName
}
