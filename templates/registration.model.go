//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"

type RegistrationTemplateModel struct {
	Config *aegis.AegisConfig
	ErrorMsg string
	LoginInfo *LoginInfoModel
}

