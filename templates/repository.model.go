//go:build ignore
package templates

import "github.com/bctnry/gitus/pkg/gitlib"

type RepositoryModel struct{
	RepoName string
	RepoObj *gitlib.LocalGitRepository
	RepoHeaderInfo RepoHeaderTemplateModel
	BranchList map[string]*gitlib.Branch
	TagList map[string]*gitlib.Tag
}
