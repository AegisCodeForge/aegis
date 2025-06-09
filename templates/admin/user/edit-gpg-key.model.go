//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type SettingEditGPGKeyTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	Key *model.AegisSigningKey
	ErrorMsg struct{
		Type string
		Message string
	}
}

