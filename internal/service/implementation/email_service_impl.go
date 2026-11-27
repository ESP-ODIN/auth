package implementation

import (
	"fmt"
	"net/smtp"
)

type SMTPEmailService struct {
	host        string
	port        string
	user        string
	password    string
	from        string
	frontendURL string
}

func NewSMTPEmailService(host, port, user, password, from, frontendURL string) *SMTPEmailService {
	return &SMTPEmailService{
		host:        host,
		port:        port,
		user:        user,
		password:    password,
		from:        from,
		frontendURL: frontendURL,
	}
}

func (s *SMTPEmailService) SendResetEmail(to, token string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, token)

	subject := "Reset your password"
	body := fmt.Sprintf(
		"Hello,\r\n\r\nYou requested a password reset. Click the link below to set a new password:\r\n\r\n%s\r\n\r\nThis link expires in 1 hour.\r\n\r\nIf you did not request this, please ignore this email.",
		resetLink,
	)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s",
		s.from, to, subject, body,
	)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}

	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}
