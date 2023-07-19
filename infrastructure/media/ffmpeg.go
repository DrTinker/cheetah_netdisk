package media

import (
	"NetDesk/conf"
	"fmt"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type MediaClientImpl struct {
}

func NewMediaClient() (*MediaClientImpl, error) {
	return &MediaClientImpl{}, nil
}

func (m *MediaClientImpl) GetPicThumbNail(picPath, compressPath string) (err error) {
	err = ffmpeg.Input(picPath).
		// ffmpeg -i input.jpg -vf scale=320:240 output_320x240.png
		Output(compressPath, ffmpeg.KwArgs{"y": "", "vf": "scale=" + conf.Default_ThumbNail_Scale}).
		Run()
	if err != nil {
		return err
	}
	return nil
}

func (m *MediaClientImpl) GetVideoThumbNail(videoPath, snapshotPath string, frameNum int) (err error) {
	err = ffmpeg.Input(videoPath).
		// 传入帧大于总帧数则设置为最后一帧
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
		// 截取第一帧，生成1张 名image2 编码方式为jpg
		Output(snapshotPath, ffmpeg.KwArgs{"y": "", "vframes": 1, "format": "image2", "s": conf.Default_ThumbNail_Scale}).
		Run()
	if err != nil {
		return err
	}

	return nil
}
