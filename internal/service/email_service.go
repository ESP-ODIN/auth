package service

type EmailSender interface {
	SendResetEmail(to, token string) error
}
