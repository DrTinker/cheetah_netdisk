package start

import (
	"NetDisk/client"
	"NetDisk/infrastructure/media"
)

func InitMedia() {
	impl, err := media.NewMediaClient()
	if err != nil {
		panic(err)
	}
	client.InitMediaClient(impl)
}
