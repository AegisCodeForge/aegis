//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus"
import "github.com/bctnry/gitus/pkg/gitus/model"

type SettingUserInfoTemplateModel struct {
	Config *gitus.GitusConfig
	User *model.GitusUser
	LoginInfo *LoginInfoModel
	ErrorMsg struct{
		Type string
		Message string
	}
		
}

