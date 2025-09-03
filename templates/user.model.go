//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type UserTemplateModel struct {
	Config *aegis.AegisConfig
	User *model.AegisUser
	RepositoryList []*model.Repository
	LoginInfo *LoginInfoModel
	BelongingNamespaceList []*model.Namespace
	PageInfo *PageInfoModel
	Query string
}

