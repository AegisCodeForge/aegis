//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type SettingUserInfoTemplateModel struct {
	Config *aegis.AegisConfig
	User *model.AegisUser
	LoginInfo *LoginInfoModel
	ErrorMsg struct{
		Type string
		Message string
	}
		
}

