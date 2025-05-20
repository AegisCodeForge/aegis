//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type AllNamespaceListModel struct {
	Config *aegis.AegisConfig
	DepotName string
	NamespaceList map[string]*model.Namespace
	LoginInfo *LoginInfoModel
}

