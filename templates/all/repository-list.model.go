//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type AllRepositoryListModel struct {
	Config *aegis.AegisConfig
	RepositoryList map[string]*model.Repository
	DepotName string
	LoginInfo *LoginInfoModel
}

