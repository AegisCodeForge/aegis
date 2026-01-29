//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"

type UploadFileTemplateModel struct {
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	PermaLink string

	FullTreePath string
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel
	
	LoginInfo *LoginInfoModel
}

