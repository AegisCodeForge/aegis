package smtp_plain

import (
	gomail "github.com/wneessen/go-mail"
)

type AegisSMTPPlainMailerInterface struct {
	username string
	appPassword string
	client *gomail.Client
}

func NewAegisSMTPPlainMailerInterface(host string, port int, auth string, username string, appPassword string) (*AegisSMTPPlainMailerInterface, error) {
	authType := gomail.SMTPAuthAutoDiscover
	if len(auth) > 0 { authType = gomail.SMTPAuthType(auth) }
	cl, err := gomail.NewClient(host, gomail.WithPort(port), gomail.WithSMTPAuth(authType), gomail.WithUsername(username), gomail.WithPassword(appPassword))
	if err != nil { return nil, err }
	return &AegisSMTPPlainMailerInterface{
		username: username,
		appPassword: appPassword,
		client: cl,
	}, nil
}

func (mi *AegisSMTPPlainMailerInterface) SendPlainTextMail(target string, title string, body string) error {
	msg := gomail.NewMsg()
	if err := msg.From(mi.username); err != nil { return err }
	if err := msg.To(target); err != nil { return err }
	msg.Subject(title)
	msg.SetBodyString(gomail.TypeTextPlain, body)
	return mi.client.DialAndSend(msg)
}

func (mi *AegisSMTPPlainMailerInterface) SendHTMLMail(target string, title string, body string) error {
	msg := gomail.NewMsg()
	if err := msg.From(mi.username); err != nil { return err }
	if err := msg.To(target); err != nil { return err }
	msg.Subject(title)
	msg.SetBodyString(gomail.TypeTextHTML, body)
	return mi.client.DialAndSend(msg)
}


