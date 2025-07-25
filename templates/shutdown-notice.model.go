//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"

type ShutdownNoticeTemplateModel struct{
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	Message string
}
