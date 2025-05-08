//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus"
import "github.com/bctnry/gitus/pkg/gitus/model"

type NewRepositoryTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	ErrorMsg struct{
		Type string
		Message string
	}
	NamespaceList map[string]*model.Namespace
	Selected string
}

