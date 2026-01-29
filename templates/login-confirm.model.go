//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"

type LoginConfirmTemplateModel struct {
	Config *gitus.GitusConfig
	ErrorMsg string
	LoginInfo *LoginInfoModel
	Username string
}

