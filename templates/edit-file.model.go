//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"

type EditFileTemplateModel struct {
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	PermaLink string

	FullTreePath string
	FileContent string
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel
	
	LoginInfo *LoginInfoModel
}

