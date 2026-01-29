//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"

type NewFileTemplateModel struct {
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	PermaLink string

	TargetFilePath string
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel
	
	LoginInfo *LoginInfoModel
}

