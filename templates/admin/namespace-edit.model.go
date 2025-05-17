//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type AdminNamespaceEditTemplateModel struct {
	Config *aegis.AegisConfig
	Namespace *model.Namespace
	ErrorMsg struct {
		Type string
		Message string
	}
	LoginInfo *LoginInfoModel
}

