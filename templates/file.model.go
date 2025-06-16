//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"

type FileTemplateModel struct {
	Config *aegis.AegisConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	File BlobTextTemplateModel
	PermaLink string

	TreeFileList *TreeFileListTemplateModel
	TreePath *TreePathTemplateModel
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel
	
	LoginInfo *LoginInfoModel
}

