//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"

type ErrorTemplateModel struct{
	Config *aegis.AegisConfig
	ErrorCode int
	ErrorMessage string
	LoginInfo *LoginInfoModel
}
