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
	switch cfg.Mailer.Type {
	case "gmail-plain":
		return gmail_plain.NewAegisGmailPlainMailerInterface(cfg.Mailer.User, cfg.Mailer.Password)
	case "smtp":
		return smtp_plain.NewAegisSMTPPlainMailerInterface(cfg.Mailer.SMTPServer, cfg.Mailer.SMTPPort, cfg.Mailer.SMTPAuth, cfg.Mailer.User, cfg.Mailer.Password)
	}
	return nil, ErrNotSupported
}


