//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type AdminUserListTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	ErrorMsg string
	UserList []*model.AegisUser
	PageInfo *PageInfoModel
}

