package mail

import (
	"errors"

	"github.com/bctnry/gitus/pkg/gitus"
	"github.com/bctnry/gitus/pkg/gitus/mail/gmail_plain"
)

type GitusMailerInterface interface {
	SendPlainTextMail(target string, title string, body string) error
	SendHTMLMail(target string, title string, body string) error
}

var ErrNotSupported = errors.New("Type not supported.")

func InitializeMailer(cfg *gitus.GitusConfig) (GitusMailerInterface, error) {
	switch cfg.Mailer.Type {
	case "gmail-plain":
		return gmail_plain.NewGitusGmailPlainMailerInterface(cfg.Mailer.User, cfg.Mailer.Password)
	}
	return nil, ErrNotSupported
}


