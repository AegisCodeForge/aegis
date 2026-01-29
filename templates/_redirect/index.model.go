//go:build ignore

package templates

type RedirectWithMessageModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	ErrorMsg string
	Timeout int
	RedirectUrl string
	MessageTitle string
	MessageText string
}

