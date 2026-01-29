//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"

type MaintenanceNoticeTemplateModel struct{
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	Message string
}
