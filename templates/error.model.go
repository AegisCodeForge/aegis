//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"

type ErrorTemplateModel struct{
	Config *gitus.GitusConfig
	ErrorCode int
	ErrorMessage string
	LoginInfo *LoginInfoModel
}
