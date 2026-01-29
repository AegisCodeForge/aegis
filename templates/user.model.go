//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type UserTemplateModel struct {
	Config *gitus.GitusConfig
	User *model.GitusUser
	RepositoryList []*model.Repository
	LoginInfo *LoginInfoModel
	BelongingNamespaceList []*model.Namespace
	PageInfo *PageInfoModel
	Query string
}

