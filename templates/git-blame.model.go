//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"

type GitBlameTemplateModel struct {
	Config *aegis.AegisConfig
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

