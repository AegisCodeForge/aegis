//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"

type DepotNoNamespaceModel struct {
	Config *aegis.AegisConfig
	RepositoryList []struct{RelPath string; Description string}
	DepotName string
	LoginInfo *LoginInfoModel
}

