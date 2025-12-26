//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type RepositoryNewPullRequestTemplateModel struct {
	Config *aegis.AegisConfig
	Repository *model.Repository
	RepoHeaderInfo *RepoHeaderTemplateModel
	LoginInfo *LoginInfoModel
	ErrorMsg string
	Stage string
	ReceiverBranch string
	ProviderRepository []*model.Repository
	ChosenProviderRepository *model.Repository
	ProviderBranchList []string
}

