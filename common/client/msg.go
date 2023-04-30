package client

import (
	"NetDesk/common/models"
	"sync"
)

type MsgClient interface {
	SendHTMLWithTls(cfg *models.EmailConfig, to, content, subject string) error
}

var (
	msgClient MsgClient
	MsgOnce   sync.Once
)

func GetMsgClient() MsgClient {
	return msgClient
}

func InitMsglient(client MsgClient) {
	MsgOnce.Do(
		func() {
			msgClient = client
		},
	)
}
