//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type RepositoryPullRequestListTemplateModel struct {
	Config *aegis.AegisConfig
	Repository *model.Repository
	RepoHeaderInfo *RepoHeaderTemplateModel
	LoginInfo *LoginInfoModel
	ErrorMsg string
	PullRequestList []*model.PullRequest
	PageNum int
}

