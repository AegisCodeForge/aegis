//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type ForkTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	SourceRepository *model.Repository
	ForkedRepoList []*model.Repository
	NamespaceList map[string]*model.Namespace
}

