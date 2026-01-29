//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type AllRepositoryListModel struct {
	Config *gitus.GitusConfig
	RepositoryList []*model.Repository
	DepotName string
	LoginInfo *LoginInfoModel
	PageInfo *PageInfoModel
	Query string
}

