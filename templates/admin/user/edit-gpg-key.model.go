//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type SettingEditGPGKeyTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	Key *model.GitusSigningKey
	ErrorMsg struct{
		Type string
		Message string
	}
}

