//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type AdminUserGPGKeyTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	User *model.GitusUser
	KeyList []model.GitusSigningKey
	ErrorMsg struct{
		Type string
		Message string
	}
		
}

