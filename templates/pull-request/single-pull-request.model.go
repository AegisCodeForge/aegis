//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type RepositorySinglePullRequestTemplateModel struct {
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo *RepoHeaderTemplateModel
	LoginInfo *LoginInfoModel
	ErrorMsg string
	PullRequest *model.PullRequest
	PullRequestEventList []*model.PullRequestEvent
	PageNum int64
}

