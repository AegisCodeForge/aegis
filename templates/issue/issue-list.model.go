//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type RepositoryIssueListTemplateModel struct {
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo *RepoHeaderTemplateModel
	LoginInfo *LoginInfoModel
	ErrorMsg string
	IssueList []*model.Issue
	PageInfo *PageInfoModel
	Query string
	FilterType int
}

