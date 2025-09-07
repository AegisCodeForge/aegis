//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"

type LoginConfirmTemplateModel struct {
	Config *aegis.AegisConfig
	ErrorMsg string
	LoginInfo *LoginInfoModel
	Username string
}

