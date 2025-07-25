//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"

type MaintenanceNoticeTemplateModel struct{
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	Message string
}
