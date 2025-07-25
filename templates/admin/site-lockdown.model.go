//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type AdminSiteLockdownTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	CurrentMode string
	PrivateNoticeMessage string
	ShutdownNoticeMessage string
	FullAccessUser string
	MaintenanceNoticeMessage string
}

