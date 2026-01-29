package smtp_plain

import (
	gomail "github.com/wneessen/go-mail"
)

type GitusSMTPPlainMailerInterface struct {
	username string
	appPassword string
	client *gomail.Client
}

func NewGitusSMTPPlainMailerInterface(host string, port int, auth string, username string, appPassword string) (*GitusSMTPPlainMailerInterface, error) {
	authType := gomail.SMTPAuthAutoDiscover
	if len(auth) > 0 { authType = gomail.SMTPAuthType(auth) }
	cl, err := gomail.NewClient(host, gomail.WithPort(port), gomail.WithSMTPAuth(authType), gomail.WithUsername(username), gomail.WithPassword(appPassword))
	if err != nil { return nil, err }
	return &GitusSMTPPlainMailerInterface{
		username: username,
		appPassword: appPassword,
		client: cl,
	}, nil
}

func (mi *GitusSMTPPlainMailerInterface) SendPlainTextMail(target string, title string, body string) error {
	msg := gomail.NewMsg()
	if err := msg.From(mi.username); err != nil { return err }
	if err := msg.To(target); err != nil { return err }
	msg.Subject(title)
	msg.SetBodyString(gomail.TypeTextPlain, body)
	return mi.client.DialAndSend(msg)
}

func (mi *GitusSMTPPlainMailerInterface) SendHTMLMail(target string, title string, body string) error {
	msg := gomail.NewMsg()
	if err := msg.From(mi.username); err != nil { return err }
	if err := msg.To(target); err != nil { return err }
	msg.Subject(title)
	msg.SetBodyString(gomail.TypeTextHTML, body)
	return mi.client.DialAndSend(msg)
}


