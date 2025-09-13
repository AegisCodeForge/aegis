//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type NewRepositoryTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	ErrorMsg struct{
		Type string
		Message string
	}
	NamespaceList map[string]*model.Namespace
	Selected string
	PredefinedNamespace string
}

