//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"

type PrivateNoticeTemplateModel struct{
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	Message string
}
