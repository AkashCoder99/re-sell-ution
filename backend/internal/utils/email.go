package utils

import (
	"fmt"
	"net/smtp"
	"strings"
)

type EmailSender interface {
	Send(toEmail, subject, body string) error
}

type SMTPEmailSender struct {
	Host      string
	Port      string
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

func (s SMTPEmailSender) Send(toEmail, subject, body string) error {
	if strings.TrimSpace(s.Host) == "" {
		return fmt.Errorf("smtp host is not configured")
	}

	fromHeader := s.FromEmail
	if strings.TrimSpace(s.FromName) != "" {
		fromHeader = fmt.Sprintf("%s <%s>", s.FromName, s.FromEmail)
	}

	msg := []byte(
		"From: " + fromHeader + "\r\n" +
			"To: " + toEmail + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body + "\r\n",
	)

	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	var auth smtp.Auth
	if strings.TrimSpace(s.Username) != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
	}

	return smtp.SendMail(addr, auth, s.FromEmail, []string{toEmail}, msg)
}
