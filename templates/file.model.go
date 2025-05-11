//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus"

type FileTemplateModel struct {
	Config *gitus.GitusConfig
	RepoHeaderInfo RepoHeaderTemplateModel
	File BlobTextTemplateModel
	PermaLink string

	RenderedTree string
	TreePath *TreePathTemplateModel
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel
	
	LoginInfo *LoginInfoModel
}

