//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type RepositorySettingMemberListTemplateModel struct {
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo *RepoHeaderTemplateModel
	RepoFullName string
	LoginInfo *LoginInfoModel
	ErrorMsg string
	ACL map[string]*model.ACLTuple
	PageInfo *PageInfoModel
}

