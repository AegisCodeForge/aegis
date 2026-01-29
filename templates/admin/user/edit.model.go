//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type AdminUserEditTemplateModel struct {
	Config *gitus.GitusConfig
	User *model.GitusUser
	ErrorMsg struct {
		Type string
		Message string
	}
	LoginInfo *LoginInfoModel
}

