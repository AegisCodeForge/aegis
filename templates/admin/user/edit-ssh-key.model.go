//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type SettingEditSSHKeyTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	Key *model.AegisAuthKey
	ErrorMsg struct{
		Type string
		Message string
	}
}

