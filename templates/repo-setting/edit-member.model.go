//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type RepositorySettingEditMemberTemplateModel struct {
	Config *aegis.AegisConfig
	Repository *model.Repository
	RepoFullName string
	RepoHeaderInfo *RepoHeaderTemplateModel
	LoginInfo *LoginInfoModel
	ErrorMsg string
	Username string
	ACLTuple *model.ACLTuple
}

