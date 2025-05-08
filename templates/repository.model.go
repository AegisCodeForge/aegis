//go:build ignore
package templates

import "github.com/bctnry/gitus/pkg/gitus"
import "github.com/bctnry/gitus/pkg/gitlib"

type RepositoryModel struct{
	Config *gitus.GitusConfig
	RepoName string
	RepoObj *gitlib.LocalGitRepository
	RepoHeaderInfo RepoHeaderTemplateModel
	BranchList map[string]*gitlib.Branch
	TagList map[string]*gitlib.Tag
	LoginInfo *LoginInfoModel
}
