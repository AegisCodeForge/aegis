//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type AdminRepositoryListTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	ErrorMsg string
	RepositoryList []*model.Repository
	PageInfo *PageInfoModel
	Query string
}

