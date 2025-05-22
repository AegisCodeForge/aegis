//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type AdminNamespaceListTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	ErrorMsg string
	NamespaceList map[string]*model.Namespace
	PageInfo *PageInfoModel
}

