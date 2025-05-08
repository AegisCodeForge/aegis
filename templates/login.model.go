//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus"

type LoginTemplateModel struct {
	Config *gitus.GitusConfig
	ErrorMsg string
	LoginInfo *LoginInfoModel
}

