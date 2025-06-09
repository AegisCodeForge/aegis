//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type AdminUserEditTemplateModel struct {
	Config *aegis.AegisConfig
	User *model.AegisUser
	ErrorMsg struct {
		Type string
		Message string
	}
	LoginInfo *LoginInfoModel
}

