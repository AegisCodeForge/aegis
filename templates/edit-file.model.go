//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"

type EditFileTemplateModel struct {
	Config *aegis.AegisConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	PermaLink string

	FullTreePath string
	FileContent string
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel
	
	LoginInfo *LoginInfoModel
}

