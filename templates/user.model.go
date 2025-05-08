//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus"
import "github.com/bctnry/gitus/pkg/gitus/model"

type UserTemplateModel struct {
	Config *gitus.GitusConfig
	User *model.GitusUser
	RepositoryList map[string]*model.Repository
	LoginInfo *LoginInfoModel
}

