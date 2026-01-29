//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type SettingEmailTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	EmailList []struct{
		Email string
		Verified bool
		Primary bool
	}
	ErrorMsg struct{
		Type string
		Message string
	}
		
}

