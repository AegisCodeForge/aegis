//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type AdminUserEditSSHKeyTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	User *model.GitusUser
	Key *model.GitusAuthKey
	ErrorMsg struct{
		Type string
		Message string
	}
}

