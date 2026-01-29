//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"

type LoginTemplateModel struct {
	Config *gitus.GitusConfig
	ErrorMsg string
	LoginInfo *LoginInfoModel
	Callback string
}

