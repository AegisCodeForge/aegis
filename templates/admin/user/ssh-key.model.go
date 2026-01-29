//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type AdminUserSSHKeyTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	User *model.GitusUser
	KeyList []model.GitusAuthKey
	ErrorMsg struct{
		Type string
		Message string
	}
		
}

