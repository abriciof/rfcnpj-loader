package email

import (
	"crypto/tls"
	"errors"
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
	recipients, err := parseRecipients(cfg.To)
	if err != nil {
		return err
	}

	from := strings.TrimSpace(cfg.User)
	if from == "" {
		return errors.New("smtp user (from) is required")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	var auth smtp.Auth
	if strings.TrimSpace(cfg.User) != "" && strings.TrimSpace(cfg.Pass) != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	}

	msg := strings.Builder{}
	msg.WriteString("From: " + from + "\r\n")
	msg.WriteString("To: " + strings.Join(recipients, ", ") + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)
	msg.WriteString("\r\n")

	return smtp.SendMail(addr, auth, from, recipients, []byte(msg.String()))
}

func CheckConnection(cfg SMTPConfig) error {
	return checkConnection(cfg, false)
}

func CheckConnectionRequireAuth(cfg SMTPConfig) error {
	return checkConnection(cfg, true)
}

func checkConnection(cfg SMTPConfig, requireAuth bool) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	user := strings.TrimSpace(cfg.User)
	pass := strings.TrimSpace(cfg.Pass)
	hasCreds := user != "" && pass != ""

	if hasCreds || requireAuth {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: cfg.Host}); err != nil {
				return err
			}
		}

		if ok, _ := client.Extension("AUTH"); ok {
			auth := smtp.PlainAuth("", user, pass, cfg.Host)
			if err := client.Auth(auth); err != nil {
				return err
			}
		} else if requireAuth {
			return errors.New("smtp server does not support AUTH")
		}
	}

	if err := client.Noop(); err != nil {
		return err
	}
	return client.Quit()
}

func parseRecipients(to string) ([]string, error) {
	replacer := strings.NewReplacer(";", ",")
	parts := strings.Split(replacer.Replace(to), ",")

	recipients := make([]string, 0, len(parts))
	for _, p := range parts {
		addr := strings.TrimSpace(p)
		if addr == "" {
			continue
		}
		recipients = append(recipients, addr)
	}

	if len(recipients) == 0 {
		return nil, errors.New("at least one recipient is required")
	}
	return recipients, nil
}
