//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type RepositorySettingEditHookTemplateModel struct {
	Config *aegis.AegisConfig
	Repository *model.Repository
	RepoHeaderInfo *RepoHeaderTemplateModel
	RepoFullName string
	LoginInfo *LoginInfoModel
	ErrorMsg string
	HookName string
	HookSource string
}

