package mail

import (
	"errors"

	"github.com/GitusCodeForge/Gitus/pkg/gitus"
	"github.com/GitusCodeForge/Gitus/pkg/gitus/mail/gmail_plain"
	"github.com/GitusCodeForge/Gitus/pkg/gitus/mail/smtp_plain"
)

type GitusMailerInterface interface {
	SendPlainTextMail(target string, title string, body string) error
	SendHTMLMail(target string, title string, body string) error
}

var ErrNotSupported = errors.New("Type not supported.")

func InitializeMailer(cfg *gitus.GitusConfig) (GitusMailerInterface, error) {
	return CreateMailerFromMailerConfig(&cfg.Mailer)
}

func CreateMailerFromMailerConfig(cfg *gitus.GitusMailerConfig) (GitusMailerInterface, error) {
	switch cfg.Type {
	case "gmail-plain":
		return gmail_plain.NewGitusGmailPlainMailerInterface(cfg.User, cfg.Password)
	case "smtp":
		return smtp_plain.NewGitusSMTPPlainMailerInterface(cfg.SMTPServer, cfg.SMTPPort, cfg.SMTPAuth, cfg.User, cfg.Password)
	}
	return nil, ErrNotSupported
}


