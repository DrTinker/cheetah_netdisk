package object

import (
	"NetDisk/client"
	"NetDisk/conf"
)

func volumeCheck(size int64, uuid string) (now, total int64, err error) {
	now, total, err = client.GetDBClient().GetUserVolume(uuid)
	if err != nil {
		return 0, 0, err
	}
	if total < now+size {
		return now, total, conf.VolumeError
	}
	return now, total, nil
}
