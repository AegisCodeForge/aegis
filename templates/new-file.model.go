//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"

type NewFileTemplateModel struct {
	Config *aegis.AegisConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	PermaLink string

	TargetFilePath string
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel
	
	LoginInfo *LoginInfoModel
}

