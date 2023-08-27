package client

import "sync"

type MediaClient interface {
	GetPicThumbNail(picPath, compressPath string) (err error)
	GetVideoThumbNail(videoPath, snapshotPath string, frameNum int) (err error)
}

var (
	media     MediaClient
	MediaOnce sync.Once
)

func GetMediaClient() MediaClient {
	return media
}

func InitMediaClient(client MediaClient) {
	MediaOnce.Do(
		func() {
			media = client
		},
	)
}
