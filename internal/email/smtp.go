package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
	To   string
}

func Enabled(cfg SMTPConfig) bool {
	return strings.TrimSpace(cfg.User) != "" &&
		strings.TrimSpace(cfg.Pass) != "" &&
		strings.TrimSpace(cfg.To) != ""
}

func Send(cfg SMTPConfig, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)

	msg := strings.Builder{}
	msg.WriteString("From: " + cfg.User + "")
	msg.WriteString("To: " + cfg.To + "")
	msg.WriteString("Subject: " + subject + "")
	msg.WriteString("MIME-Version: 1.0")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8")
	msg.WriteString("")
	msg.WriteString(body)
	msg.WriteString("")

	return smtp.SendMail(addr, auth, cfg.User, []string{cfg.To}, []byte(msg.String()))
}
