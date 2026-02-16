package email

import (
	"feedback-app/config"
	"fmt"
	"net/smtp"
)

type Client interface {
	Send(to string, subject string, body string) error
}

type SMTPClient struct {
	cfg config.SMTPConfig
}

func NewSMTPClient(cfg config.SMTPConfig) *SMTPClient {
	return &SMTPClient{cfg: cfg}
}

func (c *SMTPClient) Send(to string, subject string, body string) error {
	var auth smtp.Auth
	if c.cfg.User != "" {
		auth = smtp.PlainAuth("", c.cfg.User, c.cfg.Password, c.cfg.Host)
	}

	headers := map[string]string{
		"From":         c.cfg.From,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=\"UTF-8\"",
	}

	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body

	addr := fmt.Sprintf("%s:%s", c.cfg.Host, c.cfg.Port)

	return smtp.SendMail(addr, auth, c.cfg.From, []string{to}, []byte(msg))
}
