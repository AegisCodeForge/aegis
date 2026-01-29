//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus"

type GitBlameTemplateModel struct {
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	Blame *gitlib.PorcelainBlame

	TreeFileList *TreeFileListTemplateModel
	TreePath *TreePathTemplateModel
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel
	PermaLink string
	
	LoginInfo *LoginInfoModel
}

