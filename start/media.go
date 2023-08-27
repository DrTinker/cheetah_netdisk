package start

import (
	"NetDesk/client"
	"NetDesk/infrastructure/media"
)

func InitMedia() {
	impl, err := media.NewMediaClient()
	if err != nil {
		panic(err)
	}
	client.InitMediaClient(impl)
}
