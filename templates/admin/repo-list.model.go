//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type AdminRepositoryListTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	ErrorMsg string
	RepositoryList []*model.Repository
	PageInfo *PageInfoModel
	Query string
}

