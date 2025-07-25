//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"

type PrivateNoticeTemplateModel struct{
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	Message string
}
