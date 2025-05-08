//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus"
import "github.com/bctnry/gitus/pkg/gitus/model"

type DepotNamespaceModel struct {
	Config *gitus.GitusConfig
	DepotName string
	NamespaceList map[string]*model.Namespace
	LoginInfo *LoginInfoModel
}

