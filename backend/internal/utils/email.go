package utils

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
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
	Security  string
	Timeout   time.Duration
}

func (s SMTPEmailSender) Send(toEmail, subject, body string) error {
	if err := s.Validate(); err != nil {
		return err
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

	client, err := s.newClient()
	if err != nil {
		return err
	}
	defer client.Close()

	if err := s.authenticate(client); err != nil {
		return err
	}
	if err := client.Mail(s.FromEmail); err != nil {
		return err
	}
	if err := client.Rcpt(toEmail); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func (s SMTPEmailSender) Validate() error {
	if strings.TrimSpace(s.Host) == "" {
		return fmt.Errorf("smtp host is not configured")
	}
	if strings.TrimSpace(s.Port) == "" {
		return fmt.Errorf("smtp port is not configured")
	}
	if _, err := mail.ParseAddress(s.FromEmail); err != nil {
		return fmt.Errorf("smtp from email is invalid: %w", err)
	}
	if strings.TrimSpace(s.Security) == "" {
		s.Security = "starttls"
	}
	switch strings.ToLower(strings.TrimSpace(s.Security)) {
	case "starttls", "tls", "none":
	default:
		return fmt.Errorf("smtp security must be starttls, tls, or none")
	}
	if (strings.TrimSpace(s.Username) == "") != (strings.TrimSpace(s.Password) == "") {
		return fmt.Errorf("smtp username and password must either both be set or both be empty")
	}
	return nil
}

func (s SMTPEmailSender) newClient() (*smtp.Client, error) {
	addr := net.JoinHostPort(s.Host, s.Port)
	timeout := s.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	security := strings.ToLower(strings.TrimSpace(s.Security))
	if security == "" {
		security = "starttls"
	}

	var (
		conn net.Conn
		err  error
	)
	dialer := &net.Dialer{Timeout: timeout}
	if security == "tls" {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
			ServerName: s.Host,
			MinVersion: tls.VersionTLS12,
		})
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return nil, err
	}

	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if security != "tls" {
		if err := s.maybeStartTLS(client, security); err != nil {
			_ = client.Close()
			return nil, err
		}
	}

	return client, nil
}

func (s SMTPEmailSender) maybeStartTLS(client *smtp.Client, security string) error {
	if security == "none" {
		return nil
	}

	ok, _ := client.Extension("STARTTLS")
	if !ok {
		return fmt.Errorf("smtp server does not advertise STARTTLS")
	}

	return client.StartTLS(&tls.Config{
		ServerName: s.Host,
		MinVersion: tls.VersionTLS12,
	})
}

func (s SMTPEmailSender) authenticate(client *smtp.Client) error {
	if strings.TrimSpace(s.Username) == "" {
		return nil
	}

	ok, _ := client.Extension("AUTH")
	if !ok {
		return fmt.Errorf("smtp server does not support AUTH")
	}

	return client.Auth(smtp.PlainAuth("", s.Username, s.Password, s.Host))
}
