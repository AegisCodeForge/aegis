//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type RepositorySettingEditMemberTemplateModel struct {
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo *RepoHeaderTemplateModel
	RepoFullName string
	LoginInfo *LoginInfoModel
	ErrorMsg string
	Username string
	ACLTuple *model.ACLTuple
}

