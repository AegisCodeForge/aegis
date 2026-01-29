//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type RepositoryNewPullRequestTemplateModel struct {
	Config *gitus.GitusConfig
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

