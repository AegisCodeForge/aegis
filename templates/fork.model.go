//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type ForkTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	SourceRepository *model.Repository
	ForkedRepoList []*model.Repository
	NamespaceList map[string]*model.Namespace
}

