//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus"

type RegistrationTemplateModel struct {
	Config *gitus.GitusConfig
	ErrorMsg string
	LoginInfo *LoginInfoModel
}

