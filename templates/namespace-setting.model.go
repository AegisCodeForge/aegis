//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus"
import "github.com/bctnry/gitus/pkg/gitus/model"

type NamespaceSettingTemplateModel struct {
	Config *gitus.GitusConfig
	DepotName string
	Namespace *model.Namespace
	LoginInfo *LoginInfoModel
}

