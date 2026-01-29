//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type AdminNamespaceEditTemplateModel struct {
	Config *gitus.GitusConfig
	Namespace *model.Namespace
	ErrorMsg struct {
		Type string
		Message string
	}
	LoginInfo *LoginInfoModel
}

