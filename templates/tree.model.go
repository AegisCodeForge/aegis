//go:build ignore
package templates

import "github.com/bctnry/aegis/pkg/aegis"

type TreeTemplateModel struct {
	Config *aegis.AegisConfig
	RepoHeaderInfo RepoHeaderTemplateModel
	TreeFileList TreeFileListTemplateModel
	PermaLink string

	TreePath *TreePathTemplateModel
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel

	LoginInfo *LoginInfoModel
}
