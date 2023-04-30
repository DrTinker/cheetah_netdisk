package msg

import (
	"NetDesk/common/models"
	"crypto/tls"
	"net/smtp"
	"strconv"

	"github.com/pkg/errors"

	"github.com/jordan-wright/email"
)

type MsgClientImpl struct {
	Email *email.Email
}

func NewMsgClientImpl() (*MsgClientImpl, error) {
	m := &MsgClientImpl{}
	e := email.NewEmail()
	m.Email = e
	return m, nil
}

func (m *MsgClientImpl) SendHTMLWithTls(cfg *models.EmailConfig, to, content, subject string) error {
	e := m.Email

	e.From = cfg.Name + " <" + cfg.Email + ">"
	e.To = []string{to}
	e.Subject = subject
	e.HTML = []byte(content)
	err := e.SendWithTLS(cfg.Address+":"+strconv.Itoa(cfg.Port), smtp.PlainAuth("", cfg.Email, cfg.Password, cfg.Address),
		&tls.Config{InsecureSkipVerify: true, ServerName: cfg.Address})

	return errors.Wrap(err, "[MsgClientImpl] SendWithTls err:")
}
