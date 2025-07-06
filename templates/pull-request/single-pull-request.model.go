//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type RepositorySinglePullRequestTemplateModel struct {
	Config *aegis.AegisConfig
	Repository *model.Repository
	RepoHeaderInfo *RepoHeaderTemplateModel
	LoginInfo *LoginInfoModel
	ErrorMsg string
	PullRequest *model.PullRequest
	PullRequestEventList []*model.PullRequestEvent
	PageNum int
}

