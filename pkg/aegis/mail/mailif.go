package mail

import (
	"errors"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/mail/gmail_plain"
	"github.com/bctnry/aegis/pkg/aegis/mail/smtp_plain"
)

type AegisMailerInterface interface {
	SendPlainTextMail(target string, title string, body string) error
	SendHTMLMail(target string, title string, body string) error
}

var ErrNotSupported = errors.New("Type not supported.")

func InitializeMailer(cfg *aegis.AegisConfig) (AegisMailerInterface, error) {
	return CreateMailerFromMailerConfig(&cfg.Mailer)
}

func CreateMailerFromMailerConfig(cfg *aegis.AegisMailerConfig) (AegisMailerInterface, error) {
	switch cfg.Type {
	case "gmail-plain":
		return gmail_plain.NewAegisGmailPlainMailerInterface(cfg.User, cfg.Password)
	case "smtp":
		return smtp_plain.NewAegisSMTPPlainMailerInterface(cfg.SMTPServer, cfg.SMTPPort, cfg.SMTPAuth, cfg.User, cfg.Password)
	}
	return nil, ErrNotSupported
}


