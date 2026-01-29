//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type SettingGPGKeyTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	KeyList []model.GitusSigningKey
	ErrorMsg struct{
		Type string
		Message string
	}
		
}

