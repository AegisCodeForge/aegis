//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type NamespaceTemplateModel struct {
	Config *gitus.GitusConfig
	Namespace *model.Namespace
	LoginInfo *LoginInfoModel
}

