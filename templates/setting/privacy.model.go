//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type SettingPrivacyTemplateModel struct {
	Config *gitus.GitusConfig
	User *model.GitusUser
	LoginInfo *LoginInfoModel
	ErrorMsg struct{
		Type string
		Message string
	}
		
}

