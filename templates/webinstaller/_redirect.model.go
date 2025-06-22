//go:build ignore

package templates

type WebInstRedirectWithMessageModel struct {
	ErrorMsg string
	Timeout int
	RedirectUrl string
	MessageTitle string
	MessageText string
}
