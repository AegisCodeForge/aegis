//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type LabelModel struct {
	Config *gitus.GitusConfig
	RepositoryList []*model.Repository
	LoginInfo *LoginInfoModel
	PageInfo *PageInfoModel
	Label string
}

