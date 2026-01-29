//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type SettingSSHKeyTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	KeyList []model.GitusAuthKey
	ErrorMsg struct{
		Type string
		Message string
	}
		
}

