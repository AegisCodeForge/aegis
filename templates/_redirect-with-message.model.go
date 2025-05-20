//go:build ignore

package templates

type RedirectWithMessageModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	Timeout int
	RedirectUrl string
	MessageTitle string
	MessageText string
}
