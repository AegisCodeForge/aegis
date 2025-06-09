//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type AdminUserSSHKeyTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	User *model.AegisUser
	KeyList []model.AegisAuthKey
	ErrorMsg struct{
		Type string
		Message string
	}
		
}

