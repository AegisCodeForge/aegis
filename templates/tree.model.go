//go:build ignore
package templates

import "github.com/bctnry/gitus/pkg/gitus"

type TreeTemplateModel struct {
	Config *gitus.GitusConfig
	RepoHeaderInfo RepoHeaderTemplateModel
	TreeFileList TreeFileListTemplateModel
	PermaLink string

	TreePath *TreePathTemplateModel
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel

	LoginInfo *LoginInfoModel
}
