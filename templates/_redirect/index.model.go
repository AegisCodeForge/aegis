//go:build ignore

package templates

type RedirectWithMessageModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	ErrorMsg string
	Timeout int
	RedirectUrl string
	MessageTitle string
	MessageText string
}

