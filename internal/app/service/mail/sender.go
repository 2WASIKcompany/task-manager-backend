package mail

import (
	"net/smtp"
	"task-manager-backend/internal/app/config"
)

type Sender struct {
	auth smtp.Auth
	addr string
	from string
}

func NewSender(cfg config.ServiceConfiguration) *Sender {
	return &Sender{
		auth: smtp.PlainAuth("", cfg.From, cfg.Token, cfg.Host),
		addr: cfg.Addr,
		from: cfg.From,
	}
}

func (s *Sender) SendMail(email, msg string) error {
	return smtp.SendMail(s.addr, s.auth, s.from, []string{email}, []byte(msg))
}
