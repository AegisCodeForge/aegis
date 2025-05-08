//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus"
import "github.com/bctnry/gitus/pkg/gitus/model"

type SettingGPGKeyTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	KeyList []model.GitusAuthKey
	ErrorMsg struct{
		Type string
		Message string
	}
		
}

