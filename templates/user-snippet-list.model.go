//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type UserSnippetListTemplateModel struct {
	Config *gitus.GitusConfig
	User *model.GitusUser
	SnippetList []*model.Snippet
	LoginInfo *LoginInfoModel
	BelongingNamespaceList []*model.Namespace
	PageInfo *PageInfoModel
	Query string
}

