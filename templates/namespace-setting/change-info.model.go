//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type NamespaceSettingTemplateModel struct {
	Config *gitus.GitusConfig
	Namespace *model.Namespace
	LoginInfo *LoginInfoModel
	ErrorMsg struct {
		Type string
		Message string
	}
}

