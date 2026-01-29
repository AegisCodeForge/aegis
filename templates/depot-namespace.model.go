//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type DepotNamespaceModel struct {
	Config *gitus.GitusConfig
	DepotName string
	NamespaceList map[string]*model.Namespace
	LoginInfo *LoginInfoModel
}

