package gmail_plain

import (
	gomail "github.com/wneessen/go-mail"
)

type GitusGmailPlainMailerInterface struct {
	username string
	appPassword string
	client *gomail.Client
}

func NewGitusGmailPlainMailerInterface(username string, appPassword string) (*GitusGmailPlainMailerInterface, error) {
	cl, err := gomail.NewClient("smtp.gmail.com", gomail.WithPort(587), gomail.WithSMTPAuth(gomail.SMTPAuthPlain), gomail.WithUsername(username), gomail.WithPassword(appPassword))
	if err != nil { return nil, err }
	return &GitusGmailPlainMailerInterface{
		username: username,
		appPassword: appPassword,
		client: cl,
	}, nil
}

func (mi *GitusGmailPlainMailerInterface) SendPlainTextMail(target string, title string, body string) error {
	msg := gomail.NewMsg()
	if err := msg.From(mi.username); err != nil { return err }
	if err := msg.To(target); err != nil { return err }
	msg.Subject(title)
	msg.SetBodyString(gomail.TypeTextPlain, body)
	return mi.client.DialAndSend(msg)
}

func (mi *GitusGmailPlainMailerInterface) SendHTMLMail(target string, title string, body string) error {
	msg := gomail.NewMsg()
	if err := msg.From(mi.username); err != nil { return err }
	if err := msg.To(target); err != nil { return err }
	msg.Subject(title)
	msg.SetBodyString(gomail.TypeTextHTML, body)
	return mi.client.DialAndSend(msg)
}


