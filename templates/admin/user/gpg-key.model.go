//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type AdminUserGPGKeyTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	User *model.AegisUser
	KeyList []model.AegisSigningKey
	ErrorMsg struct{
		Type string
		Message string
	}
		
}

