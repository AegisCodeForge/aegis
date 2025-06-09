//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type AdminUserEditSSHKeyTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	User *model.AegisUser
	Key *model.AegisAuthKey
	ErrorMsg struct{
		Type string
		Message string
	}
}

