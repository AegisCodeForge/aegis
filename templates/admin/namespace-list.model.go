//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type AdminNamespaceListTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	ErrorMsg string
	NamespaceList map[string]*model.Namespace
	PageInfo *PageInfoModel
	Query string
}

