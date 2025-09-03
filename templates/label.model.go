//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type LabelModel struct {
	Config *aegis.AegisConfig
	RepositoryList []*model.Repository
	LoginInfo *LoginInfoModel
	PageInfo *PageInfoModel
	Label string
}

